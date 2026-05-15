package ui

import (
	"archive-app/handlers"

	"fyne.io/fyne/v2/app"
)

func StartUI() {
	application := app.New()
	// Создаём единственное окно и передаём его в handlers
	mainWindow := application.NewWindow("Архиватор файлов")
	handlers.SetMainWindow(mainWindow) // новая функция
	handlers.ShowLoginScreen()         // показываем экран логина
	application.Run()
}
