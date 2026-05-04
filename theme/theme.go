package theme

import "github.com/diamondburned/gotk4/pkg/gdk/v4"

// FontConfig holds all font size settings for the application
type FontConfig struct {
	HeaderSize   float64
	FilenameSize float64
	SizeTextSize float64
}

// SpacingConfig calculates dynamic spacing based on font sizes
type SpacingConfig struct {
	Fonts FontConfig
}

// NewTheme creates a new theme with default font sizes
func NewTheme() *SpacingConfig {
	return &SpacingConfig{
		Fonts: FontConfig{
			HeaderSize:   11,
			FilenameSize: 13,
			SizeTextSize: 10,
		},
	}
}

// GetFilenameYOffset returns the Y offset for filename text based on row height and font size
func (t *SpacingConfig) GetFilenameYOffset(rowHeight int) float64 {
	// Center text vertically with some padding
	return float64(rowHeight)/2 + t.Fonts.FilenameSize/3
}

// GetSizeTextYOffset returns the Y offset for size text based on row height and font size
func (t *SpacingConfig) GetSizeTextYOffset(rowHeight int) float64 {
	return float64(rowHeight)/2 + t.Fonts.SizeTextSize/3
}

// GetRowHeight returns appropriate row height based on font sizes and icon size
func (t *SpacingConfig) GetRowHeight(iconSize int) int {
	// Minimum row height based on icon size + padding
	minHeight := iconSize + 4
	// Or based on largest font + padding
	fontHeight := int(t.Fonts.FilenameSize) + 4
	if fontHeight > minHeight {
		minHeight = fontHeight
	}
	return minHeight
}

// Color theme
type ColorTheme struct {
	BackgroundColor       gdk.RGBA
	TextColor             gdk.RGBA
	SelectedBgColor       gdk.RGBA
	SelectedTextColor     gdk.RGBA
	HeaderBackgroundColor gdk.RGBA
	HeaderTextColor       gdk.RGBA
	CopyCutBgColor        gdk.RGBA
	HoverBgColor          gdk.RGBA
	AccentColor           gdk.RGBA
}

// NewColorTheme creates a new color theme with Adwaita-compatible defaults
func NewColorTheme() *ColorTheme {
	return &ColorTheme{
		BackgroundColor:       gdk.NewRGBA(0, 0, 0, 0),        // Transparent, let CSS handle it
		TextColor:             gdk.NewRGBA(1, 1, 1, 1),        // Default to white, will be overridden
		SelectedBgColor:       gdk.NewRGBA(0.2, 0.4, 0.8, 0.2), // Default accent-like
		SelectedTextColor:     gdk.NewRGBA(0.2, 0.4, 0.8, 1),   // Default accent
		HeaderBackgroundColor: gdk.NewRGBA(0, 0, 0, 0.05),     // Subtle
		HeaderTextColor:       gdk.NewRGBA(1, 1, 1, 0.6),      // Dimmed
		CopyCutBgColor:        gdk.NewRGBA(0.5, 0.5, 0.5, 0.2),
		HoverBgColor:          gdk.NewRGBA(1, 1, 1, 0.05),
		AccentColor:           gdk.NewRGBA(0.2, 0.4, 0.8, 1),
	}
}
