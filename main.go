package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)
// Функция, служащая для распаковки архива в папку "zipOutput".
func unzipFiles(fileName string){
	// Указание папки, в которую будет распакован архив.
	dst := "zipOutput"
	archive, _ := zip.OpenReader(fileName)
	// Горутина, которая закроет zip.OpenReader, по окончанию распаковки.
	defer archive.Close()

	// Перебор всех файлов в архиве и их копирование в конечную папку.
	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		dstFile, _ := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		fileInArchive, _ := f.Open()
		io.Copy(dstFile, fileInArchive)

		dstFile.Close()
		fileInArchive.Close()
	}
	// Удаление архива
	os.Remove(fileName)
}

// Функция, служащая для смены автора в core.xml
func changeAuthor (authorName string, xmlFile string) string {

	// Переменные для полей "Автор" и "Кем изменен"
	var lastModifiedBy = fmt.Sprint("<cp:lastModifiedBy>" + authorName, "</cp:lastModifiedBy>")
	var creator = fmt.Sprint("<dc:creator>" + authorName + "</dc:creator>")

	// Поиск полей в XML и изменение с помощью regexp
	creatorCng := regexp.MustCompile("<dc:creator>..*</dc:creator>")
	out := creatorCng.ReplaceAllString(xmlFile, creator)
	modifyCng := regexp.MustCompile("<cp:lastModifiedBy>..*</cp:lastModifiedBy>")
	out1 := modifyCng.ReplaceAllString(out, lastModifiedBy)

	return out1
}

func signDocument(xmlFile string) string {
	var sign = fmt.Sprint("<Company>_AuthorChanger_</Company>")

	addSign := regexp.MustCompile("<Company>.*</Company>")
	out := addSign.ReplaceAllString(xmlFile, sign)

	return out
}

// Функция, служащая для архивации папки "zipOutput" в архив.
func zipFiles(source, target string) error {
	// Создание архива
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Создание архиватора
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	base := filepath.Base(source)

	// Циклический перебор содержимого папки и добавление в архив
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if source == path {
				return nil
			}
			path += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = path[len(base)+1:]
		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		return err
	}
	if err = archive.Flush(); err != nil {
		return err
	}
	return nil
}

func main()  {
	var authorName 		= ""
	var fileName 		= os.Args[1]
	var fileNameSplit 	= strings.Split(fileName, ".")
	var AuthorChanged 	= fmt.Sprint("AuthorChanged." + fileNameSplit[1])
	var zipFile 		= fmt.Sprint(fileNameSplit[0] + ".zip")
	var coreFile 		= "zipOutput/docProps/core.xml"
	var appFile 		= "zipOutput/docProps/app.xml"

	// Получение желаемого имени автора
	fmt.Printf("Привет! В файле " + fileName + " имя автора на: ")
	authorName, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	authorName = strings.TrimRightFunc(authorName, func(c rune) bool { return c == '\r' || c == '\n' })

	unzipFiles(fileName)
	os.Remove(fileName)

	// Копирование содержимого фалов core.xml и app.xml в переменные
	coreText, _ := ioutil.ReadFile(coreFile)
	var coreXML = string(coreText)
	os.Remove(coreFile)

	appText, _ := ioutil.ReadFile(appFile)
	var appXML = string(appText)
	os.Remove(appFile)

	var	changedApp = signDocument(appXML)
	var changedCore = changeAuthor(authorName, coreXML)

	// Запись содержимого переменных в файлы core.xml и app.xml
	ioutil.WriteFile(appFile, []byte(changedApp), 0644)
	ioutil.WriteFile(coreFile, []byte(changedCore), 0644)

	// Удаление временных файлов и формирование документа
	defer os.RemoveAll("./zipOutput")
	zipFiles("zipOutput", zipFile)
	os.Rename(zipFile, AuthorChanged)

}

