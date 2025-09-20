package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed *
var assets embed.FS

var PlayerSprite = mustLoadImage("images/player.png")
var TitleFont = mustLoadFontFace("fonts/title.ttf")
var ScoreFont = mustLoadFontFace("fonts/score.ttf")
var LevelFont = mustLoadFontFace("fonts/score.ttf")
var MeteorSprites = mustLoadImages("images/meteors/*.png")
var MeteorSpritesSmall = mustLoadImages("images/meteors-small/*.png")
var LaserSprite = mustLoadImage("images/laser.png")
var ExplosionSprite = mustLoadImage("images/explosion.png")
var ExplosionSmallSprite = mustLoadImage("images/explosion-small.png")
var Explosion = createExplosion()
var ThurstSound = mustLoadOggVorbis("audio/thrust.ogg")
var ExhaustSprite = mustLoadImage("images/fire.png")
var LaserOneSound = mustLoadOggVorbis("audio/fire.ogg")
var LaserTwoSound = mustLoadOggVorbis("audio/fire.ogg")
var LaserThreeSound = mustLoadOggVorbis("audio/fire.ogg")
var ExplosionSound = mustLoadOggVorbis("audio/explosion.ogg")
var BeatOneSound = mustLoadOggVorbis("audio/beat1.ogg")
var BeatTwoSound = mustLoadOggVorbis("audio/beat2.ogg")
var LifeIndicator = mustLoadImage("images/life-indicator.png")
var ShieldSound = mustLoadOggVorbis("audio/shield.ogg")
var ShieldSprite = mustLoadImage("images/shield.png")
var ShieldIndicator = mustLoadImage("images/shield-indicator.png")
var HyperspaceIndicator = mustLoadImage("images/hyperspace.png")

func mustLoadOggVorbis(name string) *vorbis.Stream {
	f, err := assets.ReadFile(name)
	if err != nil {
		panic(err)
	}
	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(f))
	if err != nil {
		panic(err)
	}

	return stream
}

func createExplosion() []*ebiten.Image {
	var frames []*ebiten.Image
	for i := 0; i <= 11; i++ {
		frame := mustLoadImage(fmt.Sprintf("images/explosion/%d.png", i+1))
		frames = append(frames, frame)
	}

	return frames
}

func mustLoadImages(path string) []*ebiten.Image {
	matches, err := fs.Glob(assets, path)
	if err != nil {
		panic(err)
	}

	images := make([]*ebiten.Image, len(matches))
	for i, match := range matches {
		images[i] = mustLoadImage(match)
	}

	return images
}

func mustLoadFontFace(name string) *text.GoTextFaceSource {
	f, err := assets.ReadFile(name)
	if err != nil {
		panic(err)
	}

	// reader
	r := bytes.NewReader(f)

	ts, err := text.NewGoTextFaceSource(r)
	if err != nil {
		panic(err)
	}

	return ts
}

func mustLoadImage(name string) *ebiten.Image {
	f, err := assets.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}
