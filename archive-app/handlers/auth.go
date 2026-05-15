package handlers

import (
	"archive-app/db"
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var mainWindow fyne.Window
var currentUserID int

func SetMainWindow(w fyne.Window) {
	mainWindow = w
	mainWindow.Resize(fyne.NewSize(400, 300))
}

// ShowLoginScreen показывает экран входа в том же окне
func ShowLoginScreen() {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Имя пользователя")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Пароль")

	loginBtn := widget.NewButton("Войти", func() {
		userID, err := db.LoginUser(usernameEntry.Text, passwordEntry.Text)
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		currentUserID = userID
		ShowMainScreen() // переключаем на главный экран
	})

	registerBtn := widget.NewButton("Зарегистрироваться", func() {
		ShowRegisterScreen() // переключаем на форму регистрации
	})

	form := widget.NewForm(
		widget.NewFormItem("Имя пользователя", usernameEntry),
		widget.NewFormItem("Пароль", passwordEntry),
	)

	content := container.NewVBox(
		widget.NewLabel("Добро пожаловать в архиватор"),
		form,
		loginBtn,
		registerBtn,
	)
	mainWindow.SetContent(content)
	mainWindow.Show()
}

// ShowRegisterScreen показывает форму регистрации в том же окне
func ShowRegisterScreen() {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Имя пользователя")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Пароль")

	confirmEntry := widget.NewPasswordEntry()
	confirmEntry.SetPlaceHolder("Повторите пароль")

	backBtn := widget.NewButton("← Назад", func() {
		ShowLoginScreen()
	})

	registerBtn := widget.NewButton("Зарегистрироваться", func() {
		if passwordEntry.Text != confirmEntry.Text {
			dialog.ShowError(errors.New("пароли не совпадают"), mainWindow)
			return
		}
		err := db.RegisterUser(usernameEntry.Text, passwordEntry.Text)
		if err != nil {
			dialog.ShowError(err, mainWindow)
			return
		}
		dialog.ShowInformation("Успех", "Пользователь зарегистрирован! Теперь войдите.", mainWindow)
		ShowLoginScreen()
	})

	form := widget.NewForm(
		widget.NewFormItem("Имя пользователя", usernameEntry),
		widget.NewFormItem("Пароль", passwordEntry),
		widget.NewFormItem("Повтор пароля", confirmEntry),
	)

	content := container.NewVBox(
		widget.NewLabel("Регистрация нового пользователя"),
		form,
		registerBtn,
		backBtn,
	)
	mainWindow.SetContent(content)
}
