package goasteroids

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"g-asteroids/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/solarlune/resolv"
)

const (
	baseMeteorVelocity   = 0.25                    // The base speed for meteors.
	meteorSpawnTime      = 100 * time.Millisecond  // How long before meteors spawn.
	meteorSpeedUpAmount  = 0.1                     // How much do we speed a meteor up when it's timer runs out.
	meteorSpeedUpTime    = 1000 * time.Millisecond // How long to wait to speed up meteors.
	cleanUpExplosionTime = 200 * time.Millisecond
	baseBeatWaitTime     = 1600
	numberOfStars        = 1000
	alienAttackTime      = 3 * time.Second  // How long between alien attacks.
	alienSpawnTime       = 12 * time.Second // How long between alien spawns.
	baseAlienVelocity    = 0.5
)

// GameScene is the overall type for a game scene (e.g. TitleScene, GameScene, etc.).
type GameScene struct {
	player               *Player         // The player.
	baseVelocity         float64         // The base velocity for items in the game.
	meteorCount          int             // The counter for meteors.
	meteorSpawnTimer     *Timer          // The timer for spawning meteors.
	meteors              map[int]*Meteor // A map of meteors.
	meteorsForLevel      int             // # of meteors for a level.
	velocityTimer        *Timer          // The timer used for speeding up meteors.
	space                *resolv.Space   // The space for all collision objects.
	lasers               map[int]*Laser
	laserCount           int
	score                int
	explosionSmallSprite *ebiten.Image
	explosionSprite      *ebiten.Image
	explosionFrames      []*ebiten.Image
	cleanUpTimer         *Timer
	playerIsDead         bool
	audioContext         *audio.Context
	thrustPlayer         *audio.Player
	exhaust              *Exhaust
	laserOnePlayer       *audio.Player
	laserTwoPlayer       *audio.Player
	laserThreePlayer     *audio.Player
	explosionPlayer      *audio.Player
	beatOnePlayer        *audio.Player
	beatTwoPlayer        *audio.Player
	beatTimer            *Timer
	beatWaitTime         int
	playBeatOne          bool
	stars                []*Star
	currentLevel         int                 // The current level the player is
	shield               *Shield             // The players shield
	shieldsUpPlayer      *audio.Player       // The player for the shields up
	alienAttackTimer     *Timer              // The timer for allowing aliens to attack.
	alienCount           int                 // The count of aliens. We only allow one, but might change that.
	alienLaserCount      int                 // A count of alien lasers in play; used as index for map alienLasers.
	alienLaserPlayer     *audio.Player       // The audio player for alien laser sounds.
	alienLasers          map[int]*AlienLaser // A map of alien lasers currently active.
	alienSoundPlayer     *audio.Player       // The audio player for our alien sounds.
	alienSpawnTimer      *Timer              // The timer used to spawn aliens.
	aliens               map[int]*Alien      // A map of aliens.
}

// NewGameScene is a factory method for producing a new game. It's called once,
// when game play starts (and again when game play restarts).
func NewGameScene() *GameScene {
	g := &GameScene{
		meteorSpawnTimer:     NewTimer(meteorSpawnTime),
		baseVelocity:         baseMeteorVelocity,
		velocityTimer:        NewTimer(meteorSpeedUpTime),
		meteors:              make(map[int]*Meteor),
		meteorCount:          0,
		meteorsForLevel:      2,
		space:                resolv.NewSpace(ScreenWidth, ScreenHeight, 16, 16),
		lasers:               make(map[int]*Laser),
		laserCount:           0,
		explosionSprite:      assets.ExplosionSprite,
		explosionSmallSprite: assets.ExplosionSmallSprite,
		cleanUpTimer:         NewTimer(cleanUpExplosionTime),
		beatTimer:            NewTimer(2 * time.Second),
		beatWaitTime:         baseBeatWaitTime,
		currentLevel:         1,
		aliens:               make(map[int]*Alien),
		alienCount:           0,
		alienLasers:          make(map[int]*AlienLaser),
		alienLaserCount:      0,
		alienSpawnTimer:      NewTimer(alienSpawnTime),
		alienAttackTimer:     NewTimer(alienAttackTime),
	}
	g.player = NewPlayer(g)
	g.space.Add(g.player.playerObj)
	g.stars = GenerateStars(numberOfStars)

	g.explosionFrames = assets.Explosion

	// Load Audio.
	g.audioContext = audio.NewContext(48000)
	thrustPlayer, _ := g.audioContext.NewPlayer(assets.ThurstSound)
	g.thrustPlayer = thrustPlayer

	laserOnePlayer, _ := g.audioContext.NewPlayer(assets.LaserOneSound)
	g.laserOnePlayer = laserOnePlayer

	laserTwoPlayer, _ := g.audioContext.NewPlayer(assets.LaserTwoSound)
	g.laserTwoPlayer = laserTwoPlayer

	laserThreePlayer, _ := g.audioContext.NewPlayer(assets.LaserThreeSound)
	g.laserThreePlayer = laserThreePlayer

	explosionPlayer, _ := g.audioContext.NewPlayer(assets.ExplosionSound)
	g.explosionPlayer = explosionPlayer

	beatOnePlayer, _ := g.audioContext.NewPlayer(assets.BeatOneSound)
	beatTwoPlayer, _ := g.audioContext.NewPlayer(assets.BeatTwoSound)
	g.beatOnePlayer = beatOnePlayer
	g.beatTwoPlayer = beatTwoPlayer

	shieldsUpPlayer, _ := g.audioContext.NewPlayer(assets.ShieldSound)
	g.shieldsUpPlayer = shieldsUpPlayer

	alienLaserPlayer, _ := g.audioContext.NewPlayer(assets.AlienLaserSound)
	g.alienLaserPlayer = alienLaserPlayer

	alienSoundPlayer, _ := g.audioContext.NewPlayer(assets.AlienSound)
	alienSoundPlayer.SetVolume(0.5)
	g.alienSoundPlayer = alienSoundPlayer

	return g
}

// Update updates all game scene elements for the next draw. It's called once per tick.
func (g *GameScene) Update(state *State) error {
	g.player.Update()

	g.updateExhaust()

	g.updateShield()

	g.isPlayerDying()

	g.isPlayerDead(state)

	g.spawnMeteors()
	// Spawn aliens.
	g.spawnAliens()

	// Update aliens.
	for _, a := range g.aliens {
		a.Update()
	}

	// Let aliens attack (and play alien sound).
	g.letAliensAttack()

	// Update alien lasers.
	for _, al := range g.alienLasers {
		al.Update()
	}
	for _, m := range g.meteors {
		m.Update()
	}

	for _, l := range g.lasers {
		l.Update()
	}

	g.speedUpMeteors()

	g.isPlayerCollidingWithMeteor()

	g.isMeteorHitByPlayerLaser()

	// Check for player collision with alien.
	g.isPlayerCollidingWithAlien()

	// Check for alien laser collision with player.
	g.isPlayerHitByAlienLaser()

	// Check for player laser collision with alien.
	g.isAlienHitByPlayerLaser()

	g.cleanUpMeteorsAndAliens()

	g.beatSound()

	g.isLevelComplete(state)
	// Clean up offscreen aliens.
	g.removeOffscreenAliens()

	// Clean up offscreen lasers.
	g.removeOffscreenLasers()

	return nil
}

// Draw draws all game scene elements to the screen. It's called once per frame.
func (g *GameScene) Draw(screen *ebiten.Image) {
	// Draw stars.
	for _, s := range g.stars {
		s.Draw(screen)
	}

	g.player.Draw(screen)

	if g.exhaust != nil {
		g.exhaust.Draw(screen)
	}

	if g.shield != nil {
		g.shield.Draw(screen)
	}

	// Draw meteors.
	for _, m := range g.meteors {
		m.Draw(screen)
	}

	for _, l := range g.lasers {
		l.Draw(screen)
	}

	// Draw life indicators
	if len(g.player.lifeIndicators) > 0 {
		for _, x := range g.player.lifeIndicators {
			x.Draw(screen)
		}
	}

	// Draw shield indicators
	if len(g.player.shieldIndicators) > 0 {
		for _, x := range g.player.shieldIndicators {
			x.Draw(screen)
		}
	}

	// Draw hyperspace indicator
	if g.player.hyperspaceTimer == nil || g.player.hyperspaceTimer.IsReady() {
		g.player.hyperspaceIndicator.Draw(screen)
	}

	// Draw aliens.
	for _, a := range g.aliens {
		a.Draw(screen)
	}

	// Draw alien lasers.
	for _, al := range g.alienLasers {
		al.Draw(screen)
	}

	// update and draw score
	textToDraw := fmt.Sprintf("%06d", g.score)
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, 40)
	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.ScoreFont,
		Size:   24,
	}, op)

	// update and draw high score
	if g.score >= highScore {
		highScore = g.score
	}

	// update and draw high score
	textToDraw = fmt.Sprintf("HIGH SCORE %06d", highScore)
	op = &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, 75)
	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.ScoreFont,
		Size:   16,
	}, op)

	// update and draw high score
	textToDraw = fmt.Sprintf("LEVEL %d", g.currentLevel)
	op = &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, ScreenHeight-40)
	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.LevelFont,
		Size:   16,
	}, op)

}

// Layout is necessary to satisfy interface requirements from ebiten.
func (g *GameScene) Layout(outsideWidth, outsideHeight int) (ScreeWidth, ScreenHeight int) {
	return outsideWidth, outsideHeight
}
func (g *GameScene) isPlayerCollidingWithAlien() {
	for _, a := range g.aliens {
		if a.alienObj.IsIntersecting(g.player.playerObj) {
			if !a.game.player.isShielded {
				if !a.game.explosionPlayer.IsPlaying() {
					_ = a.game.explosionPlayer.Rewind()
					a.game.explosionPlayer.Play()
				}
				a.game.player.isDying = true
			}
		}
	}
}

func (g *GameScene) isPlayerHitByAlienLaser() {
	for _, l := range g.alienLasers {
		if l.laserObj.IsIntersecting(g.player.playerObj) {
			if !g.player.isShielded {
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
				g.player.isDying = true
			}
		}
	}
}

func (g *GameScene) isAlienHitByPlayerLaser() {
	for _, a := range g.aliens {
		for _, l := range g.lasers {
			if a.alienObj.IsIntersecting(l.laserObj) {
				laserData := l.laserObj.Data().(*ObjectData)
				delete(g.alienLasers, laserData.index)
				g.space.Remove(l.laserObj)
				a.sprite = g.explosionSprite
				g.score = g.score + 50
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
			}
		}
	}
}

func (g *GameScene) letAliensAttack() {
	if len(g.aliens) > 0 {
		if !g.alienSoundPlayer.IsPlaying() {
			_ = g.alienSoundPlayer.Rewind()
			g.alienSoundPlayer.Play()
		}

		// Update the alien attack timer.
		g.alienAttackTimer.Update()

		// Is the timer reached? If so, reset the timer and attack.
		if g.alienAttackTimer.IsReady() {
			g.alienAttackTimer.Reset()

			for _, a := range g.aliens {
				bounds := a.sprite.Bounds()
				halfW := float64(bounds.Dx()) / 2
				halfH := float64(bounds.Dy()) / 2

				var degreesRadian float64

				// Is the alien intelligent?
				if !a.isIntelligent {
					// Fire in a random direction.
					degreesRadian = rand.Float64() * (math.Pi * 2)
				} else {
					// Fire with some accuracy.
					degreesRadian = math.Atan2(g.player.position.Y-a.position.Y, g.player.position.X-a.position.X)
					degreesRadian = degreesRadian - math.Pi*-0.5
				}

				r := degreesRadian

				offsetX := float64(a.sprite.Bounds().Dx() - int(halfW))
				offsetY := float64(a.sprite.Bounds().Dy() - int(halfH))

				spawnPos := Vector{
					X: a.position.X + halfW + math.Sin(r) - offsetX,
					Y: a.position.Y + halfH + math.Cos(r) - offsetY,
				}

				laser := NewAlienLaser(spawnPos, r)
				g.alienLaserCount++
				g.alienLasers[g.alienLaserCount] = laser
				if !g.alienLaserPlayer.IsPlaying() {
					_ = g.alienLaserPlayer.Rewind()
					g.alienLaserPlayer.Play()
				}
			}
		}
	}
}

func (g *GameScene) removeOffscreenLasers() {
	for i, l := range g.lasers {
		if l.position.X > ScreenWidth+200 || l.position.Y > ScreenHeight+200 || l.position.X < -200 || l.position.Y < -200 {
			g.space.Remove(l.laserObj)
			delete(g.lasers, i)
		}
	}

	for i, l := range g.alienLasers {
		if l.position.X > ScreenWidth+200 || l.position.Y > ScreenHeight+200 || l.position.X < -200 || l.position.Y < -200 {
			g.space.Remove(l.laserObj)
			delete(g.alienLasers, i)
		}
	}
}

func (g *GameScene) spawnAliens() {
	g.alienSpawnTimer.Update()
	if len(g.aliens) == 0 {
		if g.alienSpawnTimer.IsReady() {
			g.alienSpawnTimer.Reset()
			rnd := rand.Intn(100-1) + 1
			if rnd > 50 {
				a := NewAlien(baseAlienVelocity, g)
				g.space.Add(a.alienObj)
				g.alienCount++
				g.aliens[g.alienCount] = a
			}
		}
	}
}

func (g *GameScene) removeOffscreenAliens() {
	for i, a := range g.aliens {
		if a.position.X > ScreenWidth+200 || a.position.Y > ScreenHeight+200 || a.position.X < -200 || a.position.Y < -200 {
			g.space.Remove(a.alienObj)
			delete(g.aliens, i)
		}
	}
}

func (g *GameScene) updateShield() {
	if g.shield != nil {
		g.shield.Update()
	}
}

func (g *GameScene) isLevelComplete(state *State) {
	if g.meteorCount >= g.meteorsForLevel && len(g.meteors) == 0 {
		g.baseVelocity = baseMeteorVelocity
		g.currentLevel++

		if g.currentLevel%5 == 0 {
			if g.player.livesRemaining < 6 {
				g.player.livesRemaining++
				x := float64(20 + len(g.player.lifeIndicators)*50.0)
				y := 20.0
				g.player.lifeIndicators = append(g.player.lifeIndicators, NewLifeIndicator(Vector{X: x, Y: y}))
			}
		}

		g.beatWaitTime = baseBeatWaitTime

		state.SceneManager.GoToScene(&LevelStartsScene{
			game:           g,
			nextLevelTimer: NewTimer(time.Second * 2),
			stars:          GenerateStars(numberOfStars),
		})
	}
}

func (g *GameScene) beatSound() {
	g.beatTimer.Update()
	if g.beatTimer.IsReady() {
		if g.playBeatOne {
			_ = g.beatOnePlayer.Rewind()
			g.beatOnePlayer.Play()
			g.beatTimer.Reset()
		} else {
			_ = g.beatTwoPlayer.Rewind()
			g.beatTwoPlayer.Play()
			g.beatTimer.Reset()
		}
		g.playBeatOne = !g.playBeatOne

		// speed up the timer
		if g.beatWaitTime > 400 {
			g.beatWaitTime = g.beatWaitTime - 25
			g.beatTimer = NewTimer(time.Millisecond * time.Duration(g.beatWaitTime))
		}
	}
}

func (g *GameScene) updateExhaust() {
	if g.exhaust != nil {
		g.exhaust.Update()
	}
}

func (g *GameScene) isMeteorHitByPlayerLaser() {
	for _, m := range g.meteors {
		for _, l := range g.lasers {
			if m.meteorObj.IsIntersecting(l.laserObj) {
				if m.meteorObj.Tags().Has(TagSmall) {
					// Small meteor
					m.sprite = g.explosionSmallSprite
					g.score++

					if !g.explosionPlayer.IsPlaying() {
						_ = g.explosionPlayer.Rewind()
						g.explosionPlayer.Play()
					}
				} else {
					// log.Println("hit large meteor")
					// Large meteor
					oldPos := m.position

					m.sprite = g.explosionSprite

					g.score++

					if !g.explosionPlayer.IsPlaying() {
						_ = g.explosionPlayer.Rewind()
						g.explosionPlayer.Play()
					}

					numToSpawn := rand.Intn(numberOfSmallMeteorsFromLargeMeteor)
					for i := 0; i < numToSpawn; i++ {
						meteor := NewSmallMeteor(baseMeteorVelocity, g, len(m.game.meteors)-1)
						meteor.position = Vector{oldPos.X + float64(rand.Intn(100-50)+50), oldPos.Y + float64(rand.Intn(100-50)+50)}
						meteor.meteorObj.SetPosition(meteor.position.X, meteor.position.Y)
						g.space.Add(meteor.meteorObj)
						g.meteorCount++
						g.meteors[m.game.meteorCount] = meteor
					}
				}
			}
		}
	}
}

func (g *GameScene) isPlayerDying() {
	if g.player.isDying {
		g.player.dyingTimer.Update()
		if g.player.dyingTimer.IsReady() {
			g.player.dyingTimer.Reset()
			g.player.dyingCounter++
			if g.player.dyingCounter == 12 {
				g.player.isDying = false
				g.player.isDead = true
			} else if g.player.dyingCounter < 12 {
				g.player.sprite = g.explosionFrames[g.player.dyingCounter]
			} else {
				// Do nothing.
			}
		}
	}
}

func (g *GameScene) isPlayerDead(state *State) {
	if g.playerIsDead {
		g.player.livesRemaining--
		if g.player.livesRemaining == 0 {

			// New High Score?
			if g.score > originalHighScore {
				err := updateHighScore(g.score)
				if err != nil {
					log.Println(err)
				}
			}

			state.SceneManager.GoToScene(&GameOverScene{
				game:        g,
				meteors:     make(map[int]*Meteor),
				meteorCount: 5,
				stars:       GenerateStars(numberOfStars),
			})
		} else {
			score := g.score
			livesRemaining := g.player.livesRemaining
			lifeSlice := g.player.lifeIndicators[:len(g.player.lifeIndicators)-1]
			stars := g.stars
			shieldsRemaining := g.player.shieldsRemaining
			shieldIndicatorSlice := g.player.shieldIndicators

			g.Reset()

			g.player.livesRemaining = livesRemaining
			g.score = score
			g.player.lifeIndicators = lifeSlice
			g.stars = stars
			g.player.shieldsRemaining = shieldsRemaining
			g.player.shieldIndicators = shieldIndicatorSlice
		}
	}
}

// spawnMeteors creates meteors, up to the maximum for a level.
func (g *GameScene) spawnMeteors() {
	g.meteorSpawnTimer.Update()
	if g.meteorSpawnTimer.IsReady() {
		g.meteorSpawnTimer.Reset()
		if len(g.meteors) < g.meteorsForLevel && g.meteorCount < g.meteorsForLevel {
			m := NewMeteor(g.baseVelocity, g, len(g.meteors)-1)
			g.space.Add(m.meteorObj)
			g.meteorCount++
			g.meteors[g.meteorCount] = m
		}
	}
}

// speedUpMeteors makes meteors move faster over time.
func (g *GameScene) speedUpMeteors() {
	g.velocityTimer.Update()
	if g.velocityTimer.IsReady() {
		g.velocityTimer.Reset()
		g.baseVelocity += meteorSpeedUpAmount
	}
}

func (g *GameScene) isPlayerCollidingWithMeteor() {
	for _, m := range g.meteors {
		if m.meteorObj.IsIntersecting(g.player.playerObj) {
			if !g.player.isShielded {
				m.game.player.isDying = true

				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
				break
			} else {
				// Bounce the meteor.
				g.bounceMeteor(m)
			}
		}
	}
}

func (g *GameScene) bounceMeteor(m *Meteor) {
	direction := Vector{
		X: (ScreenWidth/2 - m.position.X) * -1,
		Y: (ScreenHeight/2 - m.position.Y) * -1,
	}
	normalizedDirection := direction.Normalize()
	velocity := g.baseVelocity * 1.2

	movement := Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}

	m.movement = movement
}

func (g *GameScene) cleanUpMeteorsAndAliens() {
	g.cleanUpTimer.Update()
	if g.cleanUpTimer.IsReady() {
		for i, m := range g.meteors {
			if m.sprite == g.explosionSprite || m.sprite == g.explosionSmallSprite {
				delete(g.meteors, i)
				g.space.Remove(m.meteorObj)
			}
		}

		for i, a := range g.aliens {
			if a.sprite == g.explosionSprite {
				delete(g.aliens, i)
				g.space.Remove(a.alienObj)
			}
		}
		g.cleanUpTimer.Reset()
	}
}

func (g *GameScene) Reset() {
	g.player = NewPlayer(g)
	g.meteors = make(map[int]*Meteor)
	g.meteorCount = 0
	g.lasers = make(map[int]*Laser)
	g.laserCount = 0
	g.score = 0
	g.meteorSpawnTimer.Reset()
	g.baseVelocity = baseMeteorVelocity
	g.velocityTimer.Reset()
	g.playerIsDead = false
	g.exhaust = nil
	g.space.RemoveAll()
	g.space.Add(g.player.playerObj)
	g.stars = GenerateStars(numberOfStars)
	g.player.shieldsRemaining = numberOfShields
	g.player.isShielded = false
	g.aliens = make(map[int]*Alien)
	g.alienCount = 0
	g.alienLasers = make(map[int]*AlienLaser)
	g.alienLaserCount = 0
}
