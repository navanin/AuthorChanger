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

func unzipFiles(fileName string){

	dst := "zipOutput"
	archive, _ := zip.OpenReader(fileName)
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		os.MkdirAll(filepath.Dir(filePath), os.ModePerm)

		dstFile, _ := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())

		fileInArchive, _ := f.Open()

		io.Copy(dstFile, fileInArchive)

		dstFile.Close()
		fileInArchive.Close()
	}
	os.Remove(fileName)
}

func changeAuthor (authorName string, xmlFile string) string {

	var lastModifiedBy = fmt.Sprint("<cp:lastModifiedBy>" + authorName, "</cp:lastModifiedBy>")
	var creator = fmt.Sprint("<dc:creator>" + authorName + "</dc:creator>")
	var sign = fmt.Sprint("</cp:coreProperties>\n<!-- AuthorChanged - https://github.com/ragevna/AuthorChanger -->")
	creatorCng := regexp.MustCompile("<dc:creator>..*</dc:creator><cp:keywords></cp:keywords>")
	out := creatorCng.ReplaceAllString(xmlFile, creator)
	modifyCng := regexp.MustCompile("<cp:lastModifiedBy>..*</cp:lastModifiedBy>")
	out1 := modifyCng.ReplaceAllString(out, lastModifiedBy)
	addSign := regexp.MustCompile("</cp:coreProperties>")
	out2 := addSign.ReplaceAllString(out1, sign)
	print(out2)
	return out2
}

func zipFiles(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	base := filepath.Base(source)

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
	var authorName string = ""
	var fileName = os.Args[1]
	var fileNameOnly = strings.Split(fileName, ".")
	var zipFile = fmt.Sprint(fileNameOnly[0] + ".zip")
	var xmlFile = "zipOutput/docProps/core.xml"
	fmt.Printf("Привет! В файле " + fileName + " имя автора на: ")
	authorName, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	authorName = strings.TrimRightFunc(authorName, func(c rune) bool {
		return c == '\r' || c == '\n'
	})

	unzipFiles(fileName)
	os.Remove(fileName)

	text, _ := ioutil.ReadFile(xmlFile)
	var xml = string(text)
	os.Remove(xmlFile)

	var changedXML = changeAuthor(authorName, xml)
	ioutil.WriteFile(xmlFile, []byte(changedXML), 0644)


	defer os.RemoveAll("./zipOutput")
	zipFiles("zipOutput", zipFile)
	os.Rename(zipFile, "AuthorChanged.docx" )
}