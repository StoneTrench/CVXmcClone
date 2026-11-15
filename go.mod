module github.com/StoneTrench/CVXmcClone

go 1.27.0

require (
	github.com/StoneTrench/CovarexEngine v0.1.0-alpha
	github.com/StoneTrench/go-mat-lib v1.0.0
	github.com/gen2brain/raylib-go/raylib v0.62.2
	github.com/tetratelabs/wazero v1.12.0
)

require (
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/ebitengine/purego v0.10.2 // indirect
	github.com/jupiterrider/ffi v0.8.0 // indirect
	golang.org/x/exp v0.0.0-20261113113407-e3970619a1a4 // indirect
	golang.org/x/sys v0.50.0 // indirect
)

replace (
	github.com/StoneTrench/CovarexEngine => ../CovarexEngine
	github.com/StoneTrench/go-mat-lib => ../GoMatLib
)
