package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	gravity      = 0.0 // 今回は水平方向の衝突に集中するため重力は0
)

// Ball はシミュレーション内の球体を表す構造体
type Ball struct {
	x, y   float64
	vx, vy float64
	radius float64
	mass   float64
	color  color.Color
}

type Game struct {
	ball1       Ball
	ball2       Ball
	showForce   bool      // 衝突時の力を表示するフラグ
	forceTimer  time.Time // 力のベクトルを表示するタイマー
	forcePos    float64   // 衝突位置（X座標）
	forceMag    float64   // 加わった力の大きさ（力積）
}

func NewGame() *Game {
	g := &Game{}
	g.Reset()
	return g
}

// Reset で初期状態に戻す（質量に差をつけて作用反作用の違いを見やすくする）
func (g *Game) Reset() {
	g.ball1 = Ball{
		x:      150,
		y:      300,
		vx:     4.0,
		vy:     0,
		radius: 30,
		mass:   1.0, // 軽いボール
		color:  color.RGBA{230, 90, 90, 255},
	}
	g.ball2 = Ball{
		x:      650,
		y:      300,
		vx:     -2.0,
		vy:     0,
		radius: 50,
		mass:   3.0, // 重いボール（質量3倍）
		color:  color.RGBA{70, 130, 180, 255},
	}
	g.showForce = false
}

func (g *Game) Update() error {
	// スペースキーでいつでもリセット可能
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
	}

	// ボールの移動
	g.ball1.x += g.ball1.vx
	g.ball2.x += g.ball2.vx

	// 衝突判定 (2つの中心点間の距離 < 半径の和)
	dx := g.ball2.x - g.ball1.x
	dy := g.ball2.y - g.ball1.y
	distance := math.Sqrt(dx*dx + dy*dy)
	minDist := g.ball1.radius + g.ball2.radius

	if distance < minDist {
		// めり込み防止の単純な位置修正
		overlap := minDist - distance
		g.ball1.x -= overlap * 0.5
		g.ball2.x += overlap * 0.5

		// --- 物理演算：完全弾性衝突 ---
		m1 := g.ball1.mass
		m2 := g.ball2.mass
		v1 := g.ball1.vx
		v2 := g.ball2.vx

		// 衝突後の速度を求める公式 (運動量保存の法則から導出)
		newVx1 := ((m1-m2)*v1 + 2*m2*v2) / (m1 + m2)
		newVx2 := (2*m1*v1 + (m2-m1)*v2) / (m1 + m2)

		// 作用・反作用の力（力積 = 質量 × 速度の変化量）を計算
		// どちらのボールから計算しても、絶対値は必ず同じになる（F1 = -F2）
		g.forceMag = math.Abs(m1 * (newVx1 - v1))

		// 速度の更新
		g.ball1.vx = newVx1
		g.ball2.vx = newVx2

		// エフェクト表示のトリガーを設定
		g.showForce = true
		g.forceTimer = time.Now()
		g.forcePos = g.ball1.x + g.ball1.radius // 衝突点の目安
	}

	// 衝突エフェクトのタイマー（0.5秒間表示）
	if g.showForce && time.Since(g.forceTimer) > 500*time.Millisecond {
		g.showForce = false
	}

	// 画面端での跳ね返り
	if g.ball1.x-g.ball1.radius < 0 {
		g.ball1.vx = -g.ball1.vx
	}
	if g.ball2.x+g.ball2.radius > screenWidth {
		g.ball2.vx = -g.ball2.vx
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 背景を塗りつぶし
	screen.Fill(color.RGBA{20, 20, 25, 255})

	// ボール1の描画
	vector.DrawFilledCircle(screen, float32(g.ball1.x), float32(g.ball1.y), float32(g.ball1.radius), g.ball1.color, true)
	// ボール2の描画
	vector.DrawFilledCircle(screen, float32(g.ball2.x), float32(g.ball2.y), float32(g.ball2.radius), g.ball2.color, true)

	// 作用・反作用のベクトル（矢印）を描画
	if g.showForce {
		arrowLength := float32(g.forceMag * 15) // 見やすい大きさにスケーリング
		midY := float32(screenHeight / 2)
		startX := float32(g.forcePos)

		// ボール1が受ける反作用の力 (左向きの黄色い矢印)
		vector.StrokeLine(screen, startX, midY, startX-arrowLength, midY, 4, color.RGBA{255, 215, 0, 255}, true)
		vector.DrawFilledCircle(screen, startX-arrowLength, midY, 6, color.RGBA{255, 215, 0, 255}, true)

		// ボール2が受ける作用の力 (右向きの黄色い矢印)
		vector.StrokeLine(screen, startX, midY, startX+arrowLength, midY, 4, color.RGBA{255, 215, 0, 255}, true)
		vector.DrawFilledCircle(screen, startX+arrowLength, midY, 6, color.RGBA{255, 215, 0, 255}, true)
	}

	// 画面上の文字情報
	ebiten.SetWindowTitle("Action and Reaction Simulation")
	// 簡易的な状態表示（デバッグ用途）
	println(fmt.Sprintf("Ball1 Vx: %.2f | Ball2 Vx: %.2f | Impact Force: %.2f", g.ball1.vx, g.ball2.vx, g.forceMag))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
