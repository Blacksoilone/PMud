package tui

func HelpPopupContent() PopupContent {
	return PopupContent{
		Kind:  PopupHelp,
		Title: "帮助",
		Lines: []string{
			"基础命令",
			"  look/l                         查看当前位置",
			"  go <direction>                 向指定方向移动",
			"  n/s/e/w/u/d                    快速移动",
			"  northeast/ne northwest/nw southeast/se southwest/sw",
			"  get/take <item>                拿起物品",
			"  drop <item>                    放下物品",
			"  examine/x/inspect <item>       查看物品",
			"  inventory/i                    打开背包",
			"  quest                          查看任务",
			"  quit/exit                      退出游戏",
			"",
			"界面操作",
			"  ↑/↓                           浏览命令历史",
			"  ←/→ Home/End                   移动输入光标",
			"  Ctrl+R                         强制重绘",
			"  ?/F1                           打开帮助",
			"",
			"弹窗操作",
			"  Esc                            关闭弹窗",
			"  ↑/↓ PageUp/PageDown/滚轮       滚动内容",
		},
	}
}

func InventoryPopupContent() PopupContent {
	return PopupContent{Kind: PopupInventory, Title: "背包"}
}
