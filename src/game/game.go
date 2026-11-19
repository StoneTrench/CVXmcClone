package game

import (
	"context"
	"fmt"
	"image/color"
	"log/slog"
	"path"

	host "github.com/StoneTrench/CVXmcClone/bindings/go/host"
	"github.com/StoneTrench/CVXmcClone/src/game/world"
	"github.com/StoneTrench/CovarexEngine/engine"
	"github.com/StoneTrench/CovarexEngine/engine/platform"
	"github.com/StoneTrench/CovarexEngine/engine/systems/modding"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type ModCollection struct {
	modding.HeaderCollection
	Folders []string
	Modules []api.Module
}

func (m *ModCollection) Init() *ModCollection {
	m.HeaderCollection.Init()
	return m
}

func LoadModHeaders(vfs *platform.VFS) (*ModCollection, error) {
	slog.Info("Loading mod headers")

	coll := new(ModCollection).Init()

	fullpath := "./mods"

	entries, err := vfs.ReadDir(fullpath)
	if err != nil {
		return nil, err
	}
	for _, ent := range entries {
		if ent.IsDir() {
			fldr_path := path.Join(fullpath, ent.Name())
			about_path := path.Join(fldr_path, "about.json")
			data, err := vfs.ReadFile(about_path)
			if err != nil {
				return nil, fmt.Errorf("error at %s: %w", about_path, err)
			}

			var head modding.GuestModuleHeader
			err = modding.ParseJsonHeader(data, &head)
			if err != nil {
				return nil, fmt.Errorf("error at %s: %w", about_path, err)
			}
			coll.IdMap[head.ModuleId] = uint(len(coll.Headers))
			coll.Headers = append(coll.Headers, head)
			coll.Folders = append(coll.Folders, fldr_path)
		}
	}

	issues := modding.CheckForHeaderIssues(coll.HeaderCollection)
	for _, is := range issues {
		switch is.Severity {
		case modding.DependencySeverity_Info:
			slog.Info(is.Error())
		case modding.DependencySeverity_Warning:
			slog.Warn(is.Error())
		case modding.DependencySeverity_Error:
			slog.Error(is.Error())
		}
	}
	slog.Info("Loaded mod headers")

	return coll, nil
}
func LoadModWASM(vfs *platform.VFS, coll *ModCollection, ctx context.Context, RT wazero.Runtime) error {
	for i, fldr := range coll.Folders {
		wasm_path := path.Join(fldr, "main.wasm")
		data, err := vfs.ReadFile(wasm_path)
		if err != nil {
			return fmt.Errorf("failed to open wasm file: %w", err)
		}
		comp, err := RT.CompileModule(ctx, data)
		if err != nil {
			return fmt.Errorf("failed to compile module: %w", err)
		}
		defer comp.Close(ctx)

		cnf := wazero.NewModuleConfig().
			WithStartFunctions("_initialize"). // this has to be called here manually for some reason
			WithName(coll.Headers[i].ModuleId)

		inst, err := RT.InstantiateModule(ctx, comp, cnf)
		if err != nil {
			return fmt.Errorf("failed to instantiate module: %w", err)
		}

		coll.Modules = append(coll.Modules, inst)
	}
	return nil
}

func Init(vfs *platform.VFS) error {
	var err error = nil
	err = engine.Init(vfs, "logs/")
	if err != nil {
		return fmt.Errorf("failed to initialize engine: %w", err)
	}

	slog.Info(platform.GetFormattedApplicationLabel())

	ctx := context.Background()
	RT := wazero.NewRuntime(ctx)
	err = load_api(vfs, ctx, RT)
	if err != nil {
		slog.Error(err.Error())
	}
	RegistryBlocks.RegisterPanic("engine", "unknown", world.BlockType{
		IsSolid: true,
		Color:   color.RGBA{248, 0, 248, 255},
	})

	coll, err := LoadModHeaders(vfs)
	if err != nil {
		slog.Error(err.Error())
	}
	err = LoadModWASM(vfs, coll, ctx, RT)
	if err != nil {
		slog.Error(err.Error())
	}

	err = RegistryBlocks.Freeze()
	if err != nil {
		slog.Error(err.Error())
	}
	_, _, _, err = RegistryBlocks.ApplyIdMap([]string{
		"engine:unknown",
		"me.stonetrench.core:air",
		"me.stonetrench.core:stone",
		"me.stonetrench.core:dirt",
		"me.stonetrench.core:grass",
		"me.stonetrench.core:water",
	})
	if err != nil {
		slog.Error(err.Error())
	}

	return nil
}

func ErrorToStr(err error) *string {
	if err == nil {
		return nil
	}
	s := err.Error()
	return &s
}

func load_api(vfs *platform.VFS, ctx context.Context, RT wazero.Runtime) error {
	wasi_snapshot_preview1.Instantiate(ctx, RT)
	api_bldr := new(modding.HostModuleBuilder).Init(ctx, RT, "env")

	api_bldr.Add(host.HostAPILogInfo(func(m api.Module, msg string) {
		slog.Info(fmt.Sprintf("(Module) (%s) %s", m.Name(), msg))
	}))
	api_bldr.Add(host.HostAPILogWarn(func(m api.Module, msg string) {
		slog.Warn(fmt.Sprintf("(Module) (%s) %s", m.Name(), msg))
	}))
	api_bldr.Add(host.HostAPILogError(func(m api.Module, msg string) {
		slog.Error(fmt.Sprintf("(Module) (%s) %s", m.Name(), msg))
	}))

	api_bldr.Add(host.HostAPIRegAddBlock(func(m api.Module, block host.BlockType) (string, *string) {
		IsSolid := block.Kind == host.E_Solid
		namespaceid, err := RegistryBlocks.Register(m.Name(), block.Name, world.BlockType{
			Color:   color.RGBA(block.Color),
			IsSolid: IsSolid,
		})

		return namespaceid, ErrorToStr(err)
	}))
	api_bldr.Add(host.HostAPIRegGetBlock(func(m api.Module, namespaceid string) (host.BlockType, *string) {
		b, err := RegistryBlocks.GetByIdentifier(namespaceid)
		wblock := host.BlockType{}
		if err != nil {
			return wblock, ErrorToStr(err)
		}

		wblock.Name = namespaceid
		wblock.Color = host.ColorU8(b.Color)
		var Kind host.BlockKind
		if b.IsSolid {
			Kind = host.E_Solid
		} else {
			Kind = host.E_Air
		}
		wblock.Kind = Kind
		return wblock, nil
	}))
	api_bldr.Add(host.HostAPIAppendLang(func(m api.Module, name string) *string {
		// mod := modding.GetMod(m)
		// if mod == nil {
		// 	return ErrorToStr(fmt.Errorf("failed to append language resource: cant find this mod"))
		// }
		// if len(name) != 5 {
		// 	return ErrorToStr(fmt.Errorf("failed to append language resource: name must be 5 characters long"))
		// }
		// err := translate.Append(vfs, path.Join(mod.Folder, "assets", "lang", name+".json"))
		// if err != nil {
		// 	return ErrorToStr(fmt.Errorf("failed to append language resource: %w", err))
		// }
		return nil
	}))

	return api_bldr.Instantiate(ctx)
}
