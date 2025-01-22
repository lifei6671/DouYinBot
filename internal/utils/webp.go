package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/image/webp"

	webpp "github.com/chai2010/webp"
)

// Image2Webp 将图片转为webp
// inputFile 图片字节切片（仅限gif,jpeg,png格式）
// outputFile webp图片字节切片
// 图片质量
func Image2Webp(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()
	oFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer oFile.Close()
	//解析图片
	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("decode image err:%s", err)
		return err
	}
	//转为webp
	webpBytes, err := webpp.EncodeRGBA(img, 100)

	if err != nil {
		log.Printf("encode image err:%s", err)
		return err
	}
	_, oErr := oFile.Write(webpBytes)

	return oErr
}

// IsAnimatedWebP 判断一个 WebP 是否是一个动画文件
func IsAnimatedWebP(filepath string) (bool, error) {
	// 打开文件
	file, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(bufio.NewReader(file))
	scanner.Split(bufio.ScanWords)
	var isAnimWebp = false
	var current = 0
	var limit = 6
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "ANIM") || strings.Contains(scanner.Text(), "VP8X") {
			isAnimWebp = true
			break
		}

		if current > limit {
			break
		}
		// 读取到一定的行数就不读了
		current++
	}
	if isAnimWebp {
		return true, nil
	}
	file.Seek(0, io.SeekStart)

	// 读取文件头部前 12 个字节
	header := make([]byte, 12)
	_, err = file.Read(header)
	if err != nil {
		return false, err
	}

	// 检查 WebP 文件签名
	if !bytes.Equal(header[:4], []byte("RIFF")) || !bytes.Equal(header[8:12], []byte("WEBP")) {
		return false, fmt.Errorf("not a valid WebP file")
	}

	// 读取 VP8X 块，找到动画标志
	buffer := make([]byte, 1024)
	_, err = file.Read(buffer)
	if err != nil {
		return false, err
	}

	// 搜索 VP8X 块
	for i := 0; i < len(buffer)-7; i++ {
		if bytes.Equal(buffer[i:i+4], []byte("VP8X")) {
			// VP8X 块找到后，检查动画标志（第 5 字节的第 1 位是否为 1）
			animationFlag := buffer[i+4]
			return (animationFlag & 0x02) != 0, nil
		}
	}

	// 如果未找到 VP8X 块
	return false, fmt.Errorf("VP8X block not found, possibly not an extended WebP file")
}

// ExtractFirstFrame extracts the first frame of an animated WebP file
func ExtractFirstFrame(inputPath, outputPath string) error {
	// 打开输入文件
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// 解码 WebP 文件
	img, err := webp.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode WebP file: %w", err)
	}

	// 将图像保存为新的 WebP（静态图像）
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// 使用 PNG 作为中间格式保存图像（也可以直接用其他库保存为 WebP）
	err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 95})
	if err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	fmt.Println("Successfully converted animated WebP to static WebP (saved as PNG).")
	return nil
}
