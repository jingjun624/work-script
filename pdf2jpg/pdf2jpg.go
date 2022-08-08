package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"gopkg.in/gographics/imagick.v2/imagick"
)

var pdflist []string
var pdfFilePath string
var jpgFilePath string

func main() {
	pdflist = findDir("./", 0)

	_, err := PathExists("./image")

	if err != nil {
		panic(err)
	}

	for _, v := range pdflist {
		pdfFilePath = "./pdf/" + v
		ConvertPdfToImage(pdfFilePath, 800, 1212, 200, 85)
	}
	fmt.Println("转换完成!")
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		// 创建文件夹
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
		} else {
			return true, nil
		}
	}
	return false, err
}

func findDir(dir string, num int) []string {

	fileinfo, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	// 遍历这个文件夹
	for _, fi := range fileinfo {

		// 判断是不是目录
		if fi.IsDir() {
			fmt.Println(`忽略目录：`, fi.Name())
			findDir(dir+`/`+fi.Name(), num+1)
		} else {
			fileExt := path.Ext(fi.Name())
			if fileExt == ".pdf" {
				pdflist = append(pdflist, fi.Name())
			} else if fileExt == ".PDF" {
				pdflist = append(pdflist, fi.Name())
			}
		}
	}
	return pdflist
}

//ConvertPdfToImage 转换pdf为图片格式
//@resolution:扫描精度
//@CompressionQuality:图片质量: 1~100
func ConvertPdfToImage(pdfFilePath string, pageWidth uint, pageHeight uint, resolution float64, compressionQuality uint) (err error) {

	jpgFilePath = strings.Replace(pdfFilePath, "pdf", "image", -1)

	imagick.Initialize()
	defer imagick.Terminate()
	mw := imagick.NewMagickWand()
	//defer mw.Clear()
	defer mw.Destroy()

	if err := mw.SetResolution(resolution, resolution); err != nil {
		println("扫描精度设置失败")
		return err
	}

	if err := mw.ReadImage(pdfFilePath); err != nil {
		println("文件读取失败")
		return err
	}

	var pages = int(mw.GetNumberImages())
	println("待转换文件为", pdfFilePath, "总页数:", pages)

	//裁剪会使页数增加
	addPages := 0
	path := ""
	for i := 0; i < pages; i++ {
		mw.SetIteratorIndex(i) // This being the page offset

		//压平图像，去掉alpha通道，防止JPG中的alpha变黑,用在ReadImage之后
		if err := mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_FLATTEN); err != nil {
			println("图片")
			return err
		}

		mw.SetImageFormat("jpeg")
		mw.SetImageCompression(imagick.COMPRESSION_JPEG)
		mw.SetImageCompressionQuality(compressionQuality)

		//如果width>height ,就裁剪成两张
		pWidth := mw.GetImageWidth()
		pHeight := mw.GetImageHeight()

		//需要裁剪
		if pWidth > pHeight {

			//mw.ResizeImage(pageWidth*2, pageHeight, imagick.FILTER_UNDEFINED, 1.0)
			mw.ThumbnailImage(pageWidth*2, pageHeight)

			tempImage := mw.GetImageFromMagickWand()
			leftMw := imagick.NewMagickWandFromImage(tempImage)

			//左半页
			mw.CropImage(pageWidth, pageHeight, 0, 0)
			path = jpgFilePath + strconv.Itoa(i+addPages) + ".jpeg"
			mw.WriteImage(path)

			//右半页
			leftMw.SetImageFormat("jpg")
			leftMw.SetImageCompression(imagick.COMPRESSION_JPEG)
			leftMw.SetImageCompressionQuality(compressionQuality)
			leftMw.CropImage(pageWidth, pageHeight, int(pageWidth), 0)
			addPages++
			path = jpgFilePath + strconv.Itoa(i+addPages) + ".jpeg"
			leftMw.WriteImage(path)
			leftMw.Destroy()

		} else {

			mw.ThumbnailImage(pageWidth, pageHeight)
			path = jpgFilePath + strconv.Itoa(i+addPages) + ".jpeg"
			mw.WriteImage(path)

		}

	}

	return nil
}
