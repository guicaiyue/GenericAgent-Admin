//go:build windows

package main

import "embed"

//go:embed assets/ga-admin-pets/*/spritesheet.png
var gaAdminPetAssetsFS embed.FS

//go:embed assets/ga-admin-pets/ga-navigator/spritesheet.png
var gaAdminPetSpritesheetPNG []byte
