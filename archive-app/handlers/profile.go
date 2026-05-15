package handlers

import (
	"archive-app/db"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowProfileScreen() {
	mainWindow.SetTitle("Профиль пользователя")
	mainWindow.Resize(fyne.NewSize(550, 400))

	history, err := db.GetUserHistory(currentUserID)
	if err != nil {
		dialog.ShowError(err, mainWindow)
		history = []db.ArchiveRecord{}
	}

	var listData []string
	for _, rec := range history {
		action := "Создание"
		if rec.Action == "extract" {
			action = "Распаковка"
		}
		line := fmt.Sprintf("%s | %s | %s",
			rec.CreatedAt.Format("2006-01-02 15:04:05"),
			action,
			rec.ArchiveName)
		listData = append(listData, line)
	}

	historyList := widget.NewList(
		func() int { return len(listData) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, obj fyne.CanvasObject) { obj.(*widget.Label).SetText(listData[i]) },
	)

	if len(listData) == 0 {
		historyList = widget.NewList(
			func() int { return 1 },
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(i int, obj fyne.CanvasObject) { obj.(*widget.Label).SetText("История пуста") },
		)
	}

	backBtn := widget.NewButton("← На главную", func() {
		ShowMainScreen()
	})

	content := container.NewBorder(
		widget.NewLabelWithStyle("📜 История операций", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		backBtn,
		nil,
		nil,
		container.NewScroll(historyList),
	)
	mainWindow.SetContent(content)
	mainWindow.Show()
}
