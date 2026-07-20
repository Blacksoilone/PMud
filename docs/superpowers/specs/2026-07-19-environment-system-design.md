# Environment System Design

## Purpose

This document records the agreed long-term design for the MUD environment system (temperature, humidity, light level, wind speed, weather, time, and seasons). It is a design specification, not an implementation plan.

Implementation will begin after the RPG system (player attributes, skills, stats, status effects) is in place, so players can actively respond to and interact with their environment (e.g., cold resistance, wind penalties, heat damage).

## Core Principle

The environment is a real-time computed layer on top of static room data. Each room declares baseline values. Time, weather, seasons, and player items produce offsets. The final effective value is computed on demand.

A room's environment is not stored as runtime state—it is derived each time from:

```
Effective = Baseline + SeasonOffset + TimeOffset + WeatherOffset + ItemEffects
```

This means the environment is deterministic given the current game time, weather, and item states. No per-room runtime environment map is needed.

## Terms

### GameTime

A global game clock managed by the World. Tracks hour, day, month, year. Advances by a ticker goroutine launched in Loop.Start(). All time-dependent calculations read from the single GameTime value.

### Season

Derived from the current month. Four seasons: Spring, Summer, Autumn, Winter. Each season applies a set of offsets to temperature, humidity, and baseline weather probabilities.

### WeatherState

A global weather pattern (Clear, Cloudy, Rain, Storm, Snow, Fog) with an intensity value (1–100). The WeatherState is stored on World and transitions probabilistically based on season and current weather duration.

### RoomEnvironment

A computed snapshot of a room's current environmental conditions. Returned by `World.RoomEnvironment(roomID)`. Contains Temperature, Humidity, LightLevel, WindSpeed, and Weather. This is NOT stored—it is computed on demand.

### Baseline

Per-room static values declared in source.json / RoomSource. These represent the room's natural state: a desert room has high baseline temperature, a cave has low baseline light.

## Design

### GameTime

```
// internal/world/time.go

type GameTime struct {
    Hour   int   // 0–23
    Day    int   // 1–30
    Month  int   // 1–12
    Year   int
}

func (t GameTime) Season() Season
func (t GameTime) IsDaytime() bool           // typically 06–20
func (t GameTime) SunAngle() float64         // 0–90, for light calculation
```

World stores a single `gameTime GameTime` field. A goroutine in Loop.Start advances it on a configurable real-time-to-game-time ratio (e.g., 1 real minute = 1 game hour).

On each time advancement, the loop checks for weather transitions and broadcasts time-change events if needed.

### WeatherState

```
// internal/world/weather.go

type WeatherPattern int

const (
    WeatherClear  WeatherPattern = iota
    WeatherCloudy
    WeatherRain
    WeatherStorm
    WeatherSnow
    WeatherFog
)

type WeatherState struct {
    Pattern      WeatherPattern
    Intensity    int   // 1–100
    TransitionIn int   // game hours until next possible transition
}

func (w *World) tickWeather()    // called by time advancement
```

Weather transitions use a season-dependent probability table:

| Season   | Clear | Cloudy | Rain | Storm | Snow | Fog |
|----------|-------|--------|------|-------|------|-----|
| Spring   | 30%   | 30%    | 25%  | 10%   | 0%   | 5%  |
| Summer   | 50%   | 20%    | 15%  | 10%   | 0%   | 5%  |
| Autumn   | 25%   | 30%    | 25%  | 10%   | 5%   | 5%  |
| Winter   | 20%   | 25%    | 15%  | 10%   | 20%  | 10% |

### Room Baseline Fields

```
// internal/world/types.go — Room

type Room struct {
    NameKey        string
    DescriptionKey string
    Name           string
    Description    string
    Dark           bool   // replaced by LightLevel-based check in future

    // Environment baselines (added in Phase 1)
    BaseTemperature   int   // celsius × 10 (250 = 25.0°C)
    BaseHumidity      int   // 0–100
    BaseLightLevel    int   // 0–100 (100 = fully lit outdoors daytime)
    BaseWindSpeed     int   // km/h × 10 (50 = 5.0 km/h)
}
```

These exist on `RoomSource` → `ServerRoom` → `Room` following the same pipeline as `Dark`.

### RoomEnvironment Computation

```
// internal/world/environment.go

type RoomEnvironment struct {
    Temperature int             // celsius × 10
    Humidity    int             // 0–100
    LightLevel  int             // 0–100
    WindSpeed   int             // km/h × 10
    Weather     WeatherPattern
}

func (w *World) RoomEnvironment(rid RoomID) RoomEnvironment {
    room := w.rooms[rid]

    env := RoomEnvironment{
        Temperature: room.BaseTemperature,
        Humidity:    room.BaseHumidity,
        LightLevel:  room.BaseLightLevel,
        WindSpeed:   room.BaseWindSpeed,
        Weather:     w.weather.Pattern,
    }

    // Apply season offsets
    season := w.gameTime.Season()
    env.Temperature += seasonTemperatureOffset(season)
    env.Humidity += seasonHumidityOffset(season)

    // Apply time-of-day light offset
    if !w.gameTime.IsDaytime() {
        env.LightLevel = max(0, env.LightLevel - 60)
    }

    // Apply weather effects
    env.applyWeather(w.weather)

    // Apply item effects (items with tags like tag.heater, tag.fan)
    // For each item in player inventory & room:
    //   sum up item environment modifiers
    //   env.Temperature += sum

    // Clamp
    env.LightLevel = clamp(env.LightLevel, 0, 100)
    env.Humidity = clamp(env.Humidity, 0, 100)

    return env
}
```

### Items Interacting with Environment

```
// Tag definitions in tag.go

"tag.heater": {
    Fields: [
        {Name: "temperature_bonus", Type: TagFieldInt, Default: 10},
        {Name: "radius", Type: TagFieldInt, Default: 1},
    ],
}

"tag.fan": {
    Fields: [
        {Name: "wind_bonus", Type: TagFieldInt, Default: 20},
    ],
}

"tag.light_source": {
    Fields: [
        {Name: "light_bonus", Type: TagFieldInt, Default: 40},
        {Name: "fuel_consumption", Type: TagFieldInt, Default: 1},
    ],
}
```

### Integration with Existing Systems

- **Dark → LightLevel**: `RoomIsLit()` checks `RoomEnvironment().LightLevel >= threshold` instead of `room.Dark`. The `Dark` field on room becomes a static hint (or is removed, with caves simply having `BaseLightLevel: 0`).
- **Movement conditions**: Exit items gain optional environment conditions. E.g., a mountain pass requires `WindSpeed < 80` to traverse.
- **Look command**: Shows environment summary when entering a room or on explicit `look`.
- **Progression triggers**: `TriggerEnteredRoom` with extreme environments can trigger quest stages (e.g., "survive the blizzard").
- **Text rendering**: Environment values displayed via text keys like `"env.hot"`, `"env.cold"`, `"env.windy"`.

## Implementation Order

### Phase 1: Static Baselines (when implemented)

Add `BaseTemperature`, `BaseHumidity`, `BaseLightLevel`, `BaseWindSpeed` to the data pipeline:

1. `content/types.go` — RoomSource + ServerRoom
2. `content/compiler.go` — pass through
3. `content/fixture.go` — set default values on existing rooms
4. `data/tutorial/source.json` — same
5. `world/types.go` — Room fields
6. `world/constructors.go` — New() + NewFromSnapshot()
7. `world/environment.go` — RoomEnvironment computation (no time/weather yet)
8. Update `RoomIsLit()` to use LightLevel

### Phase 2: Time System (after Phase 1)

1. `world/time.go` — GameTime, Season
2. `Loop.Start()` — ticker goroutine advancing time
3. Events for time progression
4. Season offset applied in RoomEnvironment

### Phase 3: Weather System (after Phase 2)

1. `world/weather.go` — WeatherState, probability table
2. `tickWeather()` called from time advancement
3. Weather broadcast events
4. Weather effects in RoomEnvironment

### Phase 4: Item/Environment Interaction (after RPG system)

1. Tags: `tag.heater`, `tag.fan`, `tag.light_source`
2. Item effect aggregation in RoomEnvironment
3. Exit conditions based on environment
4. Player status effects from extreme environments

## Non-Goals

- Per-room microclimates that persist independently (weather is global; rooms have baselines)
- Complex weather simulation (no pressure systems, wind direction maps, or precipitation tracking)
- Client-side environment rendering (environment is server-side, rendered as text descriptions)
- Real-time update streaming to clients (environment changes are event-driven on transitions)
