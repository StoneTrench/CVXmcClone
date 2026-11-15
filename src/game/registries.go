package game

import (
	"github.com/StoneTrench/CovarexEngine/engine/resources"
	"github.com/StoneTrench/CVXmcClone/src/game/world"
)

var RegistryBlocks = new(resources.Registry[world.BlockType, world.BlockId]).Init("blocks")
