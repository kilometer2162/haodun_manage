package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	MaterialShapeSuperWide     = "超横形"
	MaterialShapeHorizontal    = "横向形"
	MaterialShapeSquare        = "似方形"
	MaterialShapeVertical      = "竖向形"
	MaterialShapeSuperVertical = "超竖形"
)

func DetermineMaterialShape(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	ratio := float64(width) / float64(height)
	ratio = math.Round(ratio*1000) / 1000

	switch {
	case ratio >= 5.0:
		return MaterialShapeSuperWide
	case ratio >= 1.0/0.9:
		return MaterialShapeHorizontal
	case ratio >= 1.0/1.1:
		return MaterialShapeSquare
	case ratio >= 1.0/1.8:
		return MaterialShapeVertical
	default:
		return MaterialShapeSuperVertical
	}
}

func FormatMaterialDimensions(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	return fmt.Sprintf("%d x %d", width, height)
}

func GenerateMaterialCode() string {
	timestamp := time.Now().Format("20060102T150405")
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return fmt.Sprintf("MAT%s", timestamp)
	}
	return fmt.Sprintf("MAT%s%s", timestamp, strings.ToUpper(hex.EncodeToString(suffix)))
}
