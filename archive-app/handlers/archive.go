package handlers

import (
	"archive-app/db"
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ------------------------------------------------------------
//  Вспомогательные функции архивации
// ------------------------------------------------------------

func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func addFolderToZip(zipWriter *zip.Writer, folderPath, baseInZip string) error {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(folderPath, entry.Name())
		relPath := filepath.Join(baseInZip, entry.Name())

		if entry.IsDir() {
			zipWriter.CreateHeader(&zip.FileHeader{
				Name:   relPath + "/",
				Method: zip.Store,
			})
			err = addFolderToZip(zipWriter, fullPath, relPath)
			if err != nil {
				return err
			}
		} else {
			file, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = relPath
			header.Method = zip.Deflate

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateArchiveFromFiles(paths []string, archivePath string, progress *widget.ProgressBar, statusLabel *widget.Label) error {
	log.Printf("Создание архива: %s из %d элементов", archivePath, len(paths))

	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл архива: %v", err)
	}
	defer archiveFile.Close()

	zipWriter := zip.NewWriter(archiveFile)
	defer zipWriter.Close()

	total := len(paths)
	for i, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("ошибка доступа к %s: %v", p, err)
		}

		if info.IsDir() {
			baseName := filepath.Base(p)
			log.Printf("Добавление папки: %s (как %s)", p, baseName)
			err = addFolderToZip(zipWriter, p, baseName)
		} else {
			log.Printf("Добавление файла: %s", p)
			err = addFileToZip(zipWriter, p)
		}
		if err != nil {
			return fmt.Errorf("ошибка добавления %s: %v", p, err)
		}

		progress.SetValue(float64(i+1) / float64(total))
		if statusLabel != nil {
			statusLabel.SetText(fmt.Sprintf("Добавлено %d из %d", i+1, total))
		}
	}

	log.Println("Архив успешно создан")
	return nil
}

func ExtractArchive(archivePath, destDir string, progress *widget.ProgressBar, statusLabel *widget.Label) error {
	log.Printf("Распаковка архива %s в %s", archivePath, destDir)

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть архив: %v", err)
	}
	defer reader.Close()

	total := len(reader.File)
	for i, file := range reader.File {
		progress.SetValue(float64(i+1) / float64(total))
		if statusLabel != nil {
			statusLabel.SetText(fmt.Sprintf("Распаковка %d из %d", i+1, total))
		}

		path := filepath.Join(destDir, file.Name)

		if !strings.HasPrefix(path, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("недопустимый путь в архиве: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		destFile, err := os.Create(path)
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return err
		}
	}
	log.Println("Распаковка завершена")
	return nil
}

// ------------------------------------------------------------
//  Главный экран (улучшенный дизайн, кнопка профиля)
// ------------------------------------------------------------

func ShowMainScreen() {
	mainWindow.Resize(fyne.NewSize(650, 550))
	mainWindow.SetTitle("Архиватор — главное меню")
	mainWindow.CenterOnScreen()

	// Верхняя панель с кнопкой профиля и выхода
	profileBtn := widget.NewButtonWithIcon("Профиль", theme.AccountIcon(), func() {
		ShowProfileScreen()
	})
	logoutBtn := widget.NewButtonWithIcon("Выйти", theme.LogoutIcon(), func() {
		currentUserID = 0
		ShowLoginScreen()
	})

	topBar := container.NewHBox(
		widget.NewLabelWithStyle("📦 Архиватор v2.0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		profileBtn,
		logoutBtn,
	)

	// Основное содержимое
	selectedItems := []string{}
	selectedList := widget.NewLabel("Выбрано: (ничего)")
	selectedList.Wrapping = fyne.TextWrapWord

	btnAddFile := widget.NewButtonWithIcon("Добавить файл", theme.FileIcon(), func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			path := reader.URI().Path()
			reader.Close()
			for _, p := range selectedItems {
				if p == path {
					return
				}
			}
			selectedItems = append(selectedItems, path)
			updateLabel(selectedList, selectedItems)
		}, mainWindow).Show()
	})

	btnAddFolder := widget.NewButtonWithIcon("Добавить папку", theme.FolderIcon(), func() {
		dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			path := list.Path()
			for _, p := range selectedItems {
				if p == path {
					return
				}
			}
			selectedItems = append(selectedItems, path)
			updateLabel(selectedList, selectedItems)
		}, mainWindow).Show()
	})

	btnClear := widget.NewButtonWithIcon("Очистить список", theme.DeleteIcon(), func() {
		selectedItems = []string{}
		updateLabel(selectedList, selectedItems)
	})

	progressBar := widget.NewProgressBar()
	statusLabel := widget.NewLabel("Готов")
	statusLabel.Alignment = fyne.TextAlignCenter

	btnCreateZip := widget.NewButtonWithIcon("Создать ZIP-архив", theme.DocumentSaveIcon(), func() {
		if len(selectedItems) == 0 {
			dialog.ShowInformation("Ошибка", "Не выбрано ни одного файла или папки", mainWindow)
			return
		}
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil || writer == nil {
				return
			}
			archivePath := writer.URI().Path()
			writer.Close()
			if !strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
				archivePath += ".zip"
			}
			go func() {
				err := CreateArchiveFromFiles(selectedItems, archivePath, progressBar, statusLabel)
				if err != nil {
					dialog.ShowError(err, mainWindow)
				} else {
					db.SaveArchiveHistory(currentUserID, archivePath, "create")
					dialog.ShowInformation("Успех", "ZIP-архив создан!", mainWindow)
				}
				progressBar.SetValue(0)
				statusLabel.SetText("Готов")
			}()
		}, mainWindow)
	})

	btnExtract := widget.NewButtonWithIcon("Распаковать ZIP-архив", theme.FolderOpenIcon(), func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			archivePath := reader.URI().Path()
			reader.Close()
			dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
				if err != nil || list == nil {
					return
				}
				destDir := list.Path()
				go func() {
					err := ExtractArchive(archivePath, destDir, progressBar, statusLabel)
					if err != nil {
						dialog.ShowError(err, mainWindow)
					} else {
						db.SaveArchiveHistory(currentUserID, archivePath, "extract")
						dialog.ShowInformation("Успех", "Архив распакован!", mainWindow)
					}
					progressBar.SetValue(0)
					statusLabel.SetText("Готов")
				}()
			}, mainWindow).Show()
		}, mainWindow).Show()
	})

	mainContent := container.NewVBox(
		widget.NewLabelWithStyle("🔧 Выберите файлы или папки", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			btnAddFile,
			btnAddFolder,
		),
		btnClear,
		selectedList,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			btnCreateZip,
			btnExtract,
		),
		progressBar,
		statusLabel,
	)

	content := container.NewBorder(topBar, nil, nil, nil, container.NewPadded(mainContent))
	mainWindow.SetContent(content)
	mainWindow.Show()
}

// updateLabel обновляет текст метки со списком выбранных элементов
func updateLabel(label *widget.Label, items []string) {
	if len(items) == 0 {
		label.SetText("Выбрано: (ничего)")
		return
	}
	var names []string
	for i, p := range items {
		if i >= 3 {
			names = append(names, "...")
			break
		}
		names = append(names, filepath.Base(p))
	}
	label.SetText(fmt.Sprintf("Выбрано (%d): %s", len(items), strings.Join(names, ", ")))
}
