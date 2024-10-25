package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

type OutputFormat int

const (
	FormatPNG OutputFormat = iota
	FormatJPG
)

// GIF disposal methods
const (
	disposalNone       = 0x01
	disposalBackground = 0x02
	disposalPrevious   = 0x03
)

// 将GIF帧转换为完整图像
func gifFrameToImage(gif *gif.GIF, frame int) *image.RGBA {
	bounds := gif.Image[frame].Bounds()
	img := image.NewRGBA(bounds)

	// 如果是第一帧，直接复制
	if frame == 0 {
		drawFrame(img, gif.Image[frame], bounds)
		return img
	}

	// 对于后续帧，需要考虑前面所有帧的叠加效果
	for i := 0; i <= frame; i++ {
		disposal := uint8(0)
		if i < len(gif.Disposal) {
			disposal = gif.Disposal[i]
		}

		switch disposal {
		case disposalNone:
			drawFrame(img, gif.Image[i], bounds)
		case disposalBackground:
			// 清除为背景色
			draw.Draw(img, bounds, image.Transparent, image.Point{}, draw.Src)
			drawFrame(img, gif.Image[i], bounds)
		case disposalPrevious:
			if i > 0 {
				// 保存当前状态
				temp := image.NewRGBA(bounds)
				draw.Draw(temp, bounds, img, bounds.Min, draw.Over)
				// 绘制新帧
				drawFrame(img, gif.Image[i], bounds)
				// 恢复之前的状态
				draw.Draw(img, bounds, temp, bounds.Min, draw.Over)
			} else {
				drawFrame(img, gif.Image[i], bounds)
			}
		default:
			drawFrame(img, gif.Image[i], bounds)
		}
	}

	return img
}

// 绘制单个帧
func drawFrame(dst *image.RGBA, src *image.Paletted, bounds image.Rectangle) {
	draw.Draw(dst, bounds, src, bounds.Min, draw.Over)
}

func main() {
	// 定义命令行参数
	inputFile := flag.String("input", "", "Input GIF file path")
	outputDir := flag.String("output", "", "Output directory for image files")
	format := flag.String("format", "png", "Output format: png or jpg")
	quality := flag.Int("quality", 90, "JPEG quality (1-100)")
	flag.Parse()

	// 检查必需参数
	if *inputFile == "" || *outputDir == "" {
		fmt.Println("Usage: gifconvert -input <gif_file> -output <output_directory> [-format <png|jpg>] [-quality <1-100>]")
		flag.PrintDefaults()
		return
	}

	// 确定输出格式
	var outputFormat OutputFormat
	switch *format {
	case "jpg", "jpeg":
		outputFormat = FormatJPG
	case "png":
		outputFormat = FormatPNG
	default:
		log.Fatalf("Unsupported format: %s", *format)
	}

	// 验证质量参数
	if outputFormat == FormatJPG && (*quality < 1 || *quality > 100) {
		log.Fatal("Quality must be between 1 and 100")
	}

	// 打开 GIF 文件
	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Error opening GIF file: %v", err)
	}
	defer file.Close()

	// 解码 GIF
	gifImg, err := gif.DecodeAll(file)
	if err != nil {
		log.Fatalf("Error decoding GIF: %v", err)
	}

	// 创建输出目录
	err = os.MkdirAll(*outputDir, 0755)
	if err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// 获取输入文件的基本名称（不含扩展名）
	baseFileName := filepath.Base(*inputFile)
	baseFileName = baseFileName[:len(baseFileName)-len(filepath.Ext(baseFileName))]

	// 处理每一帧
	for i := 0; i < len(gifImg.Image); i++ {
		// 生成完整帧图像
		frameImg := gifFrameToImage(gifImg, i)

		// 创建输出文件名
		ext := ".png"
		if outputFormat == FormatJPG {
			ext = ".jpg"
		}
		outFileName := fmt.Sprintf("%s_frame_%03d%s", baseFileName, i, ext)
		outPath := filepath.Join(*outputDir, outFileName)

		// 创建输出文件
		outFile, err := os.Create(outPath)
		if err != nil {
			log.Printf("Error creating output file %s: %v", outFileName, err)
			continue
		}

		// 根据格式保存文件
		switch outputFormat {
		case FormatPNG:
			err = png.Encode(outFile, frameImg)
		case FormatJPG:
			err = jpeg.Encode(outFile, frameImg, &jpeg.Options{Quality: *quality})
		}

		if err != nil {
			log.Printf("Error encoding frame %d: %v", i, err)
			outFile.Close()
			continue
		}

		outFile.Close()
		fmt.Printf("Saved frame %d as %s\n", i, outFileName)
	}

	fmt.Printf("Successfully converted GIF to %d image files\n", len(gifImg.Image))
}
