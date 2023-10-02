package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"archive/tar"

	"github.com/ulikunitz/xz"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	link, filename, err := GetDownloadLink()
	if err != nil {
		panic(err)
	}

	err = Download(link, filename)
	if err != nil {
		panic(err)
	}

	fmt.Println(link, filename, err)

}

func Download(link, filename string) error {
	// out, err := os.Create(filename)
	// if err != nil {
	// 	return err
	// }
	// defer out.Close()

	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// _, err = io.Copy(out, resp.Body)
	// if err != nil {
	// 	return err
	// }

	uncompressedStream, err := xz.NewReader(resp.Body)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("extractTarXz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(header.Name, 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarXz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarXz: Copy() failed: %s", err.Error())
			}
			outFile.Close()

		default:
			return fmt.Errorf(
				"ExtractTarXz: uknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}

	}

	return nil
}

func GetDownloadLink() (string, string, error) {
	res, err := http.Get("https://ziglang.org/download/")
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}
	table_body := doc.Find("table tbody")

	link_element := table_body.First().Children().Eq(6).Find("a").First()
	link, ok := link_element.Attr("href")
	filename := link_element.Text()
	if ok {
		return link, filename, nil
	}
	return "", "", errors.New("failed to get link")
}
