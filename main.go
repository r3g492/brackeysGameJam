package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var (
	gameObjects          = make(map[int]GameObject)
	nextGameObjectId int = 0
)

func main() {
	display := rl.GetCurrentMonitor()
	userMonitorWidth := rl.GetMonitorWidth(display)
	userMonitorHeight := rl.GetMonitorHeight(display)
	screenWidth := int32(userMonitorWidth)
	screenHeight := int32(userMonitorHeight)
	rl.InitWindow(screenWidth, screenHeight, "Electric.")
	rl.MaximizeWindow()
	if !rl.IsWindowFullscreen() {
		rl.ToggleFullscreen()
	}
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	rl.SetTargetFPS(60)

	buttonTexture2D := rl.LoadTexture("resources/button.png")
	startTexture2D := rl.LoadTexture("resources/start.png")
	if !startButtonScreen(buttonTexture2D, display, startTexture2D) {
		return
	}

	diamondTexture2D := rl.LoadTexture("resources/diamond.png")
	midPointX, midPointY := raylibWindowMidPoint(100, 100)
	player := Player{
		id:            0,
		texture:       diamondTexture2D,
		sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 50, Height: 50},
		position:      rl.Vector2{X: midPointX, Y: midPointY},
		color:         rl.Black,
		movementSpeed: 5,
	}
	gameObjects[0] = &player
	nextGameObjectId = 1

	// https://pixabay.com/music/trap-spinning-head-271171/
	bgm := rl.LoadSound("resources/spinning-head-271171.mp3")
	rl.PlaySound(bgm)
	gunShot := rl.LoadSound("resources/shotgun-03-38220.mp3")
	stageIdx := 0
	stageEnd := 5

	gameTimer := Timer{
		time.Now(),
		rl.Vector2{
			X: float32(rl.GetMonitorWidth(display) / 2),
			Y: float32(0),
		},
	}
	for ; stageIdx < stageEnd; stageIdx++ {
		if !rl.IsSoundPlaying(bgm) {
			rl.PlaySound(bgm)
		}

		// https://pixabay.com/sound-effects/female-vocal-321-countdown-240912/
		countdownSound := rl.LoadSound("resources/female-vocal-321-countdown-240912.mp3")
		countdown(countdownSound, display, fmt.Sprintf("stage %s", strconv.Itoa(stageIdx+1)))

		player.position.X = midPointX
		player.position.Y = midPointY

		// create enemy
		enemy1 := Enemy{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 50, Height: 50},
			position:      rl.Vector2{X: midPointX + 500, Y: midPointY + 500},
			color:         rl.Red,
			movementSpeed: 5,
			lastPlanVector: rl.Vector2{
				X: 0,
				Y: 0,
			},
		}
		gameObjects[nextGameObjectId] = &enemy1
		nextGameObjectId++

		enemy2 := Enemy{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 50, Height: 50},
			position:      rl.Vector2{X: midPointX - 500, Y: midPointY - 500},
			color:         rl.Red,
			movementSpeed: 5,
			lastPlanVector: rl.Vector2{
				X: 1,
				Y: 0,
			},
		}
		gameObjects[nextGameObjectId] = &enemy2
		nextGameObjectId++

		enemy3 := Enemy{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 50, Height: 50},
			position:      rl.Vector2{X: midPointX + 250, Y: midPointY + 250},
			color:         rl.Red,
			movementSpeed: 5,
			lastPlanVector: rl.Vector2{
				X: 0,
				Y: 0,
			},
		}
		gameObjects[nextGameObjectId] = &enemy3
		nextGameObjectId++

		for !rl.WindowShouldClose() {
			if rl.WindowShouldClose() {
				return
			}

			printYourTime(gameTimer, time.Now(), false, display)

			if hasWonStage() {
				if stageIdx >= stageEnd-1 {
					if WinScreen(buttonTexture2D, display, startTexture2D, gameTimer) {
						return
					}
				}
				break
			}

			playerMovement(&player)
			if playerDeathCheck(&player) {
				if gameOverScreen(buttonTexture2D, display, startTexture2D) {
					// restart game
					stageIdx = -1
					gameTimer.Init()
					CleanAllEnemyAndBullet()
					break
				}
				return
			}

			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				rl.PlaySound(gunShot)
				createBullet(diamondTexture2D, rl.GetMousePosition(), player)
			}
			bulletCollisionCheck()
			rl.BeginDrawing()
			rl.ClearBackground(rl.LightGray)
			EnemyPlan(player)
			MoveGameObjects()
			DrawGameObjects()
			rl.EndDrawing()
		}
	}
}

func createBullet(diamondTexture2D rl.Texture2D, mousePosition rl.Vector2, player Player) {
	dx := mousePosition.X - player.position.X
	dy := mousePosition.Y - player.position.Y
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance != 0 {
		unitX := dx / distance
		unitY := dy / distance
		bulletSpeed := float32(100)
		bulletVector := rl.Vector2{
			X: unitX * bulletSpeed,
			Y: unitY * bulletSpeed,
		}

		bullet := Bullet{
			id:            nextGameObjectId,
			texture:       diamondTexture2D,
			sourceRec:     rl.Rectangle{X: 0, Y: 0, Width: 10, Height: 10},
			position:      player.position,
			color:         rl.Yellow,
			movementSpeed: bulletSpeed,
			vector:        bulletVector,
		}
		gameObjects[nextGameObjectId] = &bullet
		nextGameObjectId++
	}
}

func playerMovement(player *Player) {
	isUpPressed := rl.IsKeyDown(rl.KeyW)
	isLeftPressed := rl.IsKeyDown(rl.KeyA)
	isDownPressed := rl.IsKeyDown(rl.KeyS)
	isRightPressed := rl.IsKeyDown(rl.KeyD)

	var movementPressedKeyCount float32 = 0
	if isUpPressed {
		movementPressedKeyCount++
	}
	if isLeftPressed {
		movementPressedKeyCount++
	}
	if isDownPressed {
		movementPressedKeyCount++
	}
	if isRightPressed {
		movementPressedKeyCount++
	}

	dividedMovementSpeed := player.movementSpeed / movementPressedKeyCount

	if isUpPressed {
		player.position.Y = player.position.Y - dividedMovementSpeed
	}
	if isLeftPressed {
		player.position.X = player.position.X - dividedMovementSpeed
	}
	if isDownPressed {
		player.position.Y = player.position.Y + dividedMovementSpeed
	}
	if isRightPressed {
		player.position.X = player.position.X + dividedMovementSpeed
	}
}

func playerDeathCheck(player *Player) bool {
	playerHitbox := rl.Rectangle{
		X:      player.position.X,
		Y:      player.position.Y,
		Width:  player.sourceRec.Width,
		Height: player.sourceRec.Height,
	}

	for _, obj := range gameObjects {
		if obj.IsEnemy() {
			enemyHitbox := obj.Hitbox()
			if rl.CheckCollisionRecs(playerHitbox, enemyHitbox) {
				return true
			}
		}
	}

	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()
	if playerHitbox.X+playerHitbox.Width < 0 || playerHitbox.X > float32(screenWidth) ||
		playerHitbox.Y+playerHitbox.Height < 0 || playerHitbox.Y > float32(screenHeight) {
		return true
	}
	return false
}

func startButtonScreen(
	buttonTexture2D rl.Texture2D,
	display int,
	startTexture2D rl.Texture2D,
) bool {
	button := Button{
		id:        -1,
		texture:   buttonTexture2D,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 220, Height: 100},
		position: rl.Vector2{
			X: float32(rl.GetMonitorWidth(display))/2 - 220/2,
			Y: float32(rl.GetMonitorHeight(display))/2 - 220/2,
		},
		color:  rl.White,
		status: 0,
	}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		mousePoint := rl.GetMousePosition()
		if button.CheckInput(mousePoint) {
			return true
		}
		rl.DrawTextureRec(
			startTexture2D,
			rl.Rectangle{X: 0, Y: 0, Width: 1600, Height: 900},
			rl.Vector2{X: float32(rl.GetMonitorWidth(display))/2 - 800, Y: float32(rl.GetMonitorHeight(display))/2 - 450},
			rl.Gray,
		)
		button.Draw()
		rl.EndDrawing()
	}
	return false
}

func gameOverScreen(
	buttonTexture2D rl.Texture2D,
	display int,
	startTexture2D rl.Texture2D,
) bool {
	button := Button{
		id:        -1,
		texture:   buttonTexture2D,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 220, Height: 100},
		position: rl.Vector2{
			X: float32(rl.GetMonitorWidth(display))/2 - 220/2,
			Y: float32(rl.GetMonitorHeight(display))/2 - 220/2,
		},
		color:  rl.Red,
		status: 0,
	}

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		mousePoint := rl.GetMousePosition()
		if button.CheckInput(mousePoint) {
			return true
		}

		rl.DrawTextureRec(
			startTexture2D,
			rl.Rectangle{X: 0, Y: 0, Width: 1600, Height: 900},
			rl.Vector2{X: float32(rl.GetMonitorWidth(display))/2 - 800, Y: float32(rl.GetMonitorHeight(display))/2 - 450},
			rl.Gray,
		)
		rl.DrawText(
			fmt.Sprintf(
				"you died. \n you must be perfect...",
			),
			int32(rl.GetMonitorWidth(display)/2-400),
			int32(rl.GetMonitorHeight(display)/2-400),
			100,
			rl.Red,
		)
		button.Draw()
		rl.EndDrawing()
	}
	return false
}

func WinScreen(
	buttonTexture2D rl.Texture2D,
	display int,
	startTexture2D rl.Texture2D,
	gameTimer Timer,
) bool {
	button := Button{
		id:        -1,
		texture:   buttonTexture2D,
		sourceRec: rl.Rectangle{X: 0, Y: 0, Width: 220, Height: 100},
		position: rl.Vector2{
			X: float32(rl.GetMonitorWidth(display))/2 - 220/2,
			Y: float32(rl.GetMonitorHeight(display))/2 - 220/2,
		},
		color:  rl.Purple,
		status: 0,
	}
	winTime := time.Now()
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		mousePoint := rl.GetMousePosition()
		if button.CheckInput(mousePoint) {
			return true
		}

		rl.DrawTextureRec(
			startTexture2D,
			rl.Rectangle{X: 0, Y: 0, Width: 1600, Height: 900},
			rl.Vector2{X: float32(rl.GetMonitorWidth(display))/2 - 800, Y: float32(rl.GetMonitorHeight(display))/2 - 450},
			rl.Gray,
		)
		rl.DrawText(
			fmt.Sprintf(
				"You've Won!",
			),
			int32(rl.GetMonitorWidth(display)/2-500),
			int32(rl.GetMonitorHeight(display)/2-300),
			100,
			rl.White,
		)
		printYourTime(gameTimer, winTime, true, display)
		button.Draw()
		rl.EndDrawing()
	}
	return false
}

func countdown(countdownSound rl.Sound, display int, stageName string) {
	beginTimer := Timer{}
	beginTimer.Init()
	rl.PlaySound(countdownSound)
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		secondsLeft := 1 - time.Since(beginTimer.gameInitTime).Seconds()

		rl.DrawText(
			fmt.Sprintf(
				"Be Ready for %s",
				stageName,
			),
			int32(rl.GetMonitorWidth(display)/2-300),
			int32(rl.GetMonitorHeight(display)/2-300),
			100,
			rl.Black,
		)
		if secondsLeft <= 0 {
			break
		}
		rl.EndDrawing()
	}
}

func hasWonStage() bool {
	for _, obj := range gameObjects {
		if obj.IsEnemy() {
			return false
		}
	}
	return true
}

func bulletCollisionCheck() {
	for bulletKey, bulletObj := range gameObjects {
		if bulletObj.IsBullet() {
			bulletHitbox := bulletObj.Hitbox()

			// Remove bullet if it's out of bounds.
			if bulletHitbox.X < 0 || bulletHitbox.Y < 0 ||
				bulletHitbox.X > 5000 || bulletHitbox.Y > 5000 {
				delete(gameObjects, bulletKey)
				continue
			}

			bulletCurPos := rl.Vector2{X: bulletHitbox.X, Y: bulletHitbox.Y}
			bulletPrevPos := bulletObj.PrevPosition()

			for enemyKey, enemyObj := range gameObjects {
				if bulletKey == enemyKey {
					continue
				}
				if enemyObj.IsEnemy() {
					enemyHitbox := enemyObj.Hitbox()
					if rl.CheckCollisionRecs(bulletHitbox, enemyHitbox) ||
						lineIntersectsRect(bulletPrevPos, bulletCurPos, enemyHitbox) {
						delete(gameObjects, bulletKey)
						delete(gameObjects, enemyKey)
						break
					}
				}
			}
		}
	}
}

func pointInRect(p rl.Vector2, rect rl.Rectangle) bool {
	return p.X >= rect.X && p.X <= rect.X+rect.Width &&
		p.Y >= rect.Y && p.Y <= rect.Y+rect.Height
}

func lineSegmentsIntersect(p, p2, q, q2 rl.Vector2) bool {
	orientation := func(a, b, c rl.Vector2) float32 {
		return (b.Y-a.Y)*(c.X-a.X) - (b.X-a.X)*(c.Y-a.Y)
	}
	o1 := orientation(p, p2, q)
	o2 := orientation(p, p2, q2)
	o3 := orientation(q, q2, p)
	o4 := orientation(q, q2, p2)

	return o1*o2 < 0 && o3*o4 < 0
}

func lineIntersectsRect(p, q rl.Vector2, rect rl.Rectangle) bool {
	if pointInRect(p, rect) || pointInRect(q, rect) {
		return true
	}

	topLeft := rl.Vector2{X: rect.X, Y: rect.Y}
	topRight := rl.Vector2{X: rect.X + rect.Width, Y: rect.Y}
	bottomRight := rl.Vector2{X: rect.X + rect.Width, Y: rect.Y + rect.Height}
	bottomLeft := rl.Vector2{X: rect.X, Y: rect.Y + rect.Height}

	if lineSegmentsIntersect(p, q, topLeft, topRight) ||
		lineSegmentsIntersect(p, q, topRight, bottomRight) ||
		lineSegmentsIntersect(p, q, bottomRight, bottomLeft) ||
		lineSegmentsIntersect(p, q, bottomLeft, topLeft) {
		return true
	}

	return false
}

func MoveGameObjects() {
	for _, obj := range gameObjects {
		obj.Move()
	}
}

func DrawGameObjects() {
	for _, obj := range gameObjects {
		obj.Draw()
	}
}

func EnemyPlan(player Player) {
	for _, obj := range gameObjects {
		if obj.IsEnemy() {
			obj.EnemyPlan(player)
		}
	}
}

func CleanAllEnemyAndBullet() {
	for _, obj := range gameObjects {
		if obj.IsEnemy() || obj.IsBullet() {
			obj.Delete()
		}
	}
}

type GameObject interface {
	Draw()
	GameObjectId() int
	IsEnemy() bool
	IsBullet() bool
	Hitbox() rl.Rectangle
	Move()
	PrevPosition() rl.Vector2
	EnemyPlan(player Player)
	Delete()
}

type Timer struct {
	gameInitTime time.Time
	position     rl.Vector2
}

func (t *Timer) Init() {
	t.gameInitTime = time.Now()
}

func printYourTime(gameTimer Timer, queryTime time.Time, won bool, display int) {
	duration := queryTime.Sub(gameTimer.gameInitTime)
	if won {
		rl.DrawText(
			fmt.Sprintf(
				"Your Record: %.0f s",
				duration.Seconds(),
			),
			int32(rl.GetMonitorWidth(display)/2+150),
			int32(rl.GetMonitorHeight(display)/2-250),
			50,
			rl.White,
		)
	} else {
		rl.DrawText(
			fmt.Sprintf(
				"Your Time: %.0f s",
				duration.Seconds(),
			),
			int32(gameTimer.position.X),
			int32(gameTimer.position.Y),
			100,
			rl.Black,
		)
	}
}

type Button struct {
	id               int
	texture          rl.Texture2D
	sourceRec        rl.Rectangle
	position         rl.Vector2
	color            rl.Color
	status           int
	buttonClickSound rl.Sound
}

func (b *Button) Draw() {
	if b.status == 0 {
		b.sourceRec.Y = 0
	} else if b.status == 1 {
		b.sourceRec.Y = 110
	} else if b.status == 2 {
		b.sourceRec.Y = 220
	}
	rl.DrawTextureRec(
		b.texture,
		b.sourceRec,
		b.position,
		b.color,
	)
}

func (b *Button) IsEnemy() bool {
	return false
}

func (b *Button) IsBullet() bool {
	return false
}

func (b *Button) CheckInput(
	mousePosition rl.Vector2,
) bool {
	if mousePosition.X >= b.position.X &&
		mousePosition.X <= b.sourceRec.Width+b.position.X &&
		mousePosition.Y >= b.position.Y &&
		mousePosition.Y <= b.sourceRec.Height+b.position.Y &&
		rl.IsMouseButtonDown(rl.MouseLeftButton) {
		b.status = 1
	} else {
		b.status = 0
	}

	if mousePosition.X >= b.position.X &&
		mousePosition.X <= b.sourceRec.Width+b.position.X &&
		mousePosition.Y >= b.position.Y &&
		mousePosition.Y <= b.sourceRec.Height+b.position.Y &&
		rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		b.status = 2
		return true
	}
	return false
}

func (b *Button) GameObjectId() int {
	return b.id
}

func (b *Button) Hitbox() rl.Rectangle {
	return b.sourceRec
}

func (b *Button) Move() {
	return
}

func (b *Button) EnemyPlan(player Player) {
	return
}

func (b *Button) PrevPosition() rl.Vector2 {
	return b.position
}

func (b *Button) Delete() {
	return
}

type Player struct {
	id            int
	texture       rl.Texture2D
	sourceRec     rl.Rectangle
	position      rl.Vector2
	color         rl.Color
	movementSpeed float32
}

func (p *Player) Draw() {
	rl.DrawTextureRec(
		p.texture,
		p.sourceRec,
		p.position,
		p.color,
	)
}

func (p *Player) GameObjectId() int {
	return p.id
}

func (p *Player) IsEnemy() bool {
	return false
}

func (p *Player) IsBullet() bool {
	return false
}

func (p *Player) Hitbox() rl.Rectangle {
	return p.sourceRec
}

func (p *Player) Move() {
}

func (p *Player) EnemyPlan(player Player) {
}

func (p *Player) PrevPosition() rl.Vector2 {
	return p.position
}

func (p *Player) Delete() {
	return
}

type Enemy struct {
	id             int
	texture        rl.Texture2D
	sourceRec      rl.Rectangle
	position       rl.Vector2
	color          rl.Color
	movementSpeed  float32
	lastPlanVector rl.Vector2
	/**
	0: stop
	1: run to player
	2: move movePlan 0: up 1: up-right 2: right 3: right-down 4: down 5: down-left 6: left 7: left-up
	3: run to player with anger
	*/
	plan int
	/**
	movePlan 0: up 1: up-right 2: right 3: right-down 4: down 5: down-left 6: left 7: left-up
	*/
	movePlan         int
	lastPlanInitTime time.Time
	lastPlanDuration time.Duration
}

func (e *Enemy) resetPlan() {
	e.plan = rand.Intn(4)

	if e.plan == 2 {
		e.movePlan = rand.Intn(8)
	} else {
		e.movePlan = 0
	}

	e.movementSpeed = float32(rand.Intn(5) + 5)
	if e.plan == 3 {
		e.movementSpeed += 10
	}

	e.lastPlanInitTime = time.Now()
	e.lastPlanDuration = time.Duration(rand.Intn(3)+1) * time.Second
}

func (e *Enemy) invokeRush() {
	e.movementSpeed = float32(rand.Intn(5) + 5)
	e.plan = 3
	e.movementSpeed += 10

	e.lastPlanInitTime = time.Now()
	e.lastPlanDuration = time.Duration(rand.Intn(3)+1) * time.Second
}

func (e *Enemy) isPlanOver() bool {
	timeSince := time.Since(e.lastPlanInitTime)
	if timeSince > e.lastPlanDuration {
		return true
	}
	return false
}

func (e *Enemy) isOutOfMonitor() bool {
	hb := e.Hitbox()
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()
	if hb.X+hb.Width < 0 || hb.X > float32(screenWidth) ||
		hb.Y+hb.Height < 0 || hb.Y > float32(screenHeight) {
		return true
	}
	return false
}

func (e *Enemy) Draw() {
	rl.DrawTextureRec(
		e.texture,
		e.sourceRec,
		e.position,
		e.color,
	)
}

func (e *Enemy) Hitbox() rl.Rectangle {
	return rl.Rectangle{
		X:      e.position.X,
		Y:      e.position.Y,
		Width:  e.sourceRec.Width,
		Height: e.sourceRec.Height,
	}
}

func (e *Enemy) GameObjectId() int {
	return e.id
}

func (e *Enemy) IsEnemy() bool {
	return true
}

func (e *Enemy) IsBullet() bool {
	return false
}

func (e *Enemy) Move() {
	e.position.X += e.lastPlanVector.X
	e.position.Y += e.lastPlanVector.Y
}

func (e *Enemy) EnemyPlan(player Player) {
	if e.isOutOfMonitor() {
		e.invokeRush()
	} else {
		if e.isPlanOver() {
			e.resetPlan()
		}
	}

	switch e.plan {
	case 0:
		e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
	case 1:
		dx := player.position.X - e.position.X
		dy := player.position.Y - e.position.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist != 0 {
			unitX := dx / dist
			unitY := dy / dist
			e.lastPlanVector = rl.Vector2{
				X: unitX * e.movementSpeed,
				Y: unitY * e.movementSpeed,
			}
		} else {
			e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
		}
	case 2:
		var direction rl.Vector2
		switch e.movePlan {
		case 0:
			direction = rl.Vector2{X: 0, Y: -1}
		case 1:
			direction = rl.Vector2{X: 1, Y: -1}
		case 2:
			direction = rl.Vector2{X: 1, Y: 0}
		case 3:
			direction = rl.Vector2{X: 1, Y: 1}
		case 4:
			direction = rl.Vector2{X: 0, Y: 1}
		case 5:
			direction = rl.Vector2{X: -1, Y: 1}
		case 6:
			direction = rl.Vector2{X: -1, Y: 0}
		case 7:
			direction = rl.Vector2{X: -1, Y: -1}
		default:
			direction = rl.Vector2{X: 0, Y: 0}
		}

		mag := float32(math.Sqrt(float64(direction.X*direction.X + direction.Y*direction.Y)))
		if mag != 0 {
			direction.X /= mag
			direction.Y /= mag
		}
		e.lastPlanVector = rl.Vector2{
			X: direction.X * e.movementSpeed,
			Y: direction.Y * e.movementSpeed,
		}
	case 3:
		dx := player.position.X - e.position.X
		dy := player.position.Y - e.position.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if dist != 0 {
			unitX := dx / dist
			unitY := dy / dist
			e.lastPlanVector = rl.Vector2{
				X: unitX * e.movementSpeed,
				Y: unitY * e.movementSpeed,
			}
		} else {
			e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
		}
	default:
		e.lastPlanVector = rl.Vector2{X: 0, Y: 0}
	}
}

func (e *Enemy) PrevPosition() rl.Vector2 {
	return e.position
}

func (e *Enemy) Delete() {
	delete(gameObjects, e.id)
	return
}

type Bullet struct {
	id            int
	texture       rl.Texture2D
	sourceRec     rl.Rectangle
	position      rl.Vector2
	color         rl.Color
	movementSpeed float32
	vector        rl.Vector2
}

func (b *Bullet) Draw() {
	rl.DrawTextureRec(
		b.texture,
		b.sourceRec,
		b.position,
		b.color,
	)
}

func (b *Bullet) Hitbox() rl.Rectangle {
	return rl.Rectangle{
		X:      b.position.X,
		Y:      b.position.Y,
		Width:  b.sourceRec.Width,
		Height: b.sourceRec.Height,
	}
}

func (b *Bullet) GameObjectId() int {
	return b.id
}

func (b *Bullet) IsEnemy() bool {
	return false
}

func (b *Bullet) IsBullet() bool {
	return true
}

func (b *Bullet) Move() {
	b.position.X += b.vector.X
	b.position.Y += b.vector.Y
}

func (b *Bullet) EnemyPlan(player Player) {
}

func (b *Bullet) PrevPosition() rl.Vector2 {
	return rl.Vector2{
		X: b.position.X - b.vector.X,
		Y: b.position.Y - b.vector.Y,
	}
}

func (b *Bullet) Delete() {
	delete(gameObjects, b.id)
	return
}

func raylibWindowMidPoint(elementWidth float32, elementHeight float32) (midX float32, midY float32) {
	display := rl.GetCurrentMonitor()
	return float32(rl.GetMonitorWidth(display))/2 - elementWidth/2, float32(rl.GetMonitorHeight(display))/2 - elementHeight/2
}
