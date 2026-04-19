package gambling

import (
	"embed"
	"io/fs"

	"github.com/GoMudEngine/GoMud/internal/items"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
)

var (
	//go:embed files/*
	files embed.FS
)

const dieItemId = 1040000
const coinItemId = 1040001
const tarotItemId = 1040002
const eightItemId = 1040003
const bottleItemId = 1040004
const cardsItemId = 1040005

func init() {

	plug := plugins.New(`gambling`, `1.0`)

	if err := plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	// Read the die script from the embedded FS and register it so the
	// scripting engine can find it without a corresponding file on disk.
	scriptBytes, err := fs.ReadFile(files, `files/datafiles/items/1040000-6_sided_die.js`)
	if err != nil {
		mudlog.Error("gambling: failed to read die script", "error", err)
	} else {
		items.RegisterItemScript(dieItemId, string(scriptBytes))
	}

	coinScript, err := fs.ReadFile(files, `files/datafiles/items/1040001-lucky_coin.js`)
	if err != nil {
		mudlog.Error("gambling: failed to read coin script", "error", err)
	} else {
		items.RegisterItemScript(coinItemId, string(coinScript))
	}

	tarotScript, err := fs.ReadFile(files, `files/datafiles/items/1040002-tarot_deck.js`)
	if err != nil {
		mudlog.Error("gambling: failed to read tarot script", "error", err)
	} else {
		items.RegisterItemScript(tarotItemId, string(tarotScript))
	}

	eightScript, err := fs.ReadFile(files, `files/datafiles/items/1040003-magic_8_ball.js`)
	if err != nil {
		mudlog.Error("gambling: failed to read 8-ball script", "error", err)
	} else {
		items.RegisterItemScript(eightItemId, string(eightScript))
	}

	bottleScript, err := fs.ReadFile(files, `files/datafiles/items/1040004-empty_bottle.js`)
	if err != nil {
		mudlog.Error("gambling: failed to read bottle script", "error", err)
	} else {
		items.RegisterItemScript(bottleItemId, string(bottleScript))
	}

	cardsScript, err := fs.ReadFile(files, `files/datafiles/items/1040005-deck_of_cards.js`)
	if err != nil {
		mudlog.Error("gambling: failed to read cards script", "error", err)
	} else {
		items.RegisterItemScript(cardsItemId, string(cardsScript))
	}
}
