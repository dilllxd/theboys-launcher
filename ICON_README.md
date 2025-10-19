# Icon File Required

You need to place a file called `icon.ico` in this directory for the build process.

## Optimal Icon Sizes:
The icon should include ALL these sizes in a single ICO file:
- **16x16** - Taskbar small icons, file explorer details view
- **32x32** - Desktop icons, file explorer icons
- **48x48** - Large icons view
- **256x256** - High DPI displays, modern Windows versions

## Design Recommendations:
- **Simple design** - complex details get lost at small sizes
- **High contrast** - ensures visibility at all sizes
- **TheBoys theme** - consistent with launcher branding
- **Square format** - Windows icons are always square

## Creation Tools:
- **Online**: favicon.io, convertio.co, icoconvert.com (PNG to ICO)
- **Desktop**: GIMP (with ICO export plugin), Photoshop (with ICO plugin)
- **Windows**: Visual Studio's icon editor, Greenfish Icon Editor Pro
- **Free**: IcoFX (free version), Inkscape (with export extension)

## Workflow:
1. Create a **256x256 PNG** with high-quality design
2. Use an online converter or tool to generate multi-size ICO
3. Test the icon at different sizes to ensure it's readable
4. Save as `icon.ico` in this directory

Place the finished `icon.ico` file in the same directory as this README.