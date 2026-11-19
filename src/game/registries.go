package game

import (
	"github.com/StoneTrench/CVXmcClone/src/game/world"
	"github.com/StoneTrench/CovarexEngine/engine/resources"
)

var RegistryBlocks = new(resources.Registry[world.BlockType, world.BlockId]).Init("blocks")
