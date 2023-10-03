package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"archive/tar"
	"github.com/ulikunitz/xz"
)

func main() {
	link, err := GetDownloadLink("master", "riscv64-linux")
	if err != nil {
		panic(err)
	}

	err = Download(link)
	if err != nil {
		panic(err)
	}
}

func Download(link string) error {
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

// func GetDownloadLink() (string, string, error) {
// 	res, err := http.Get("https://ziglang.org/download/")
// 	if err != nil {
// 		return "", "", err
// 	}
// 	defer res.Body.Close()

// 	if res.StatusCode != 200 {
// 		return "", "", fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
// 	}

// 	doc, err := goquery.NewDocumentFromReader(res.Body)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	table_body := doc.Find("table tbody")

// 	link_element := table_body.First().Children().Eq(6).Find("a").First()
// 	link, ok := link_element.Attr("href")
// 	filename := link_element.Text()
// 	if ok {
// 		return link, filename, nil
// 	}
// 	return "", "", errors.New("failed to get link")
// }

func GetDownloadLink(version, platform string) (string, error) {
	resp, err := http.Get("https://ziglang.org/download/index.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %s, %d", resp.Status, resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	var version_doc versionDocument
	decoder.Decode(&version_doc)
	platform_info, ok := version_doc[version][platform].(map[string]any)
	if !ok {
		return "", fmt.Errorf("could not find tarball for version: %s and platform: %s", version, platform)
	}
	link := platform_info["tarball"].(string)
	return link, nil
}

type versionDocument = map[string]versionInfo
type versionInfo = map[string]any
