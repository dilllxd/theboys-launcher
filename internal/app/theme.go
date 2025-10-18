package app

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme provides a custom theme for Winterpack Launcher
type CustomTheme struct{}

// Color returns a custom color for the theme
func (t *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x0B, G: 0x57, B: 0xD0, A: 0xFF} // Blue 700 for dark
		}
		return color.NRGBA{R: 0x19, G: 0x76, B: 0xD2, A: 0xFF} // Blue 600 for light

	case theme.ColorNameButton:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x45, G: 0x5A, B: 0x64, A: 0xFF} // Blue Gray 700 for dark
		}
		return color.NRGBA{R: 0x60, G: 0x7D, B: 0x8B, A: 0xFF} // Blue Gray 500 for light

	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xFF} // Dark gray background
		}
		return color.NRGBA{R: 0xFA, G: 0xFA, B: 0xFA, A: 0xFF} // Light gray background

	case theme.ColorNameInputBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF} // Dark surface
		}
		return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // Light surface

	case theme.ColorNameForeground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // White text
		}
		return color.NRGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xFF} // Dark text

	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF} // Green 500

	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xFF, G: 0x98, B: 0x00, A: 0xFF} // Orange 600

	case theme.ColorNameError:
		return color.NRGBA{R: 0xF4, G: 0x43, B: 0x36, A: 0xFF} // Red 500

	case theme.ColorNameHover:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0x08} // Light hover overlay

	case theme.ColorNameSeparator:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x1F} // 12% white
		}
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x12} // 7% black

	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns a custom font for the theme
func (t *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns a custom icon for the theme
func (t *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns a custom size for the theme
func (t *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 6
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 8
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return 6
	case theme.SizeNameSelectionRadius:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}