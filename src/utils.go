package main

import (
	"math"
	"sync/atomic"
	"time"
)

type U8Color struct {
	R, G, B, A uint8
}

type F32Color struct {
	R, G, B, A float32
}

type IntVec2 struct {
	X, Y int
}

type F32Vec2 struct {
	X, Y float32
}

type IntRect struct {
	X, Y, W, H int
}

type F32Rect struct {
	X, Y, W, H float32
}

type IntSize struct {
	W, H int
}

func packColor(color U8Color) uint32 {
	rgb24 := uint32(color.R)
	rgb24 = (rgb24 << 8) | uint32(color.G)
	rgb24 = (rgb24 << 8) | uint32(color.B)
	return rgb24
}

func unpackColor(color uint32) U8Color {
	return U8Color{
		R: uint8((color >> 16) & 0xff),
		G: uint8((color >> 8) & 0xff),
		B: uint8(color & 0xff),
		A: 255,
	}
}

func (c U8Color) toF32() F32Color {
	return F32Color{
		R: float32(c.R) / 256,
		G: float32(c.G) / 256,
		B: float32(c.B) / 256,
		A: float32(c.A) / 256,
	}
}

func ortho(top, left, right, bottom, near, far float32) [16]float32 {
	rml, tmb, fmn := (right - left), (top - bottom), (far - near)
	return [16]float32{
		float32(2. / rml), 0, 0, 0, // 1
		0, float32(2. / tmb), 0, 0, // 2
		0, 0, float32(-2. / fmn), 0, // 3
		float32(-(right + left) / rml), // 4
		float32(-(top + bottom) / tmb),
		float32(-(far + near) / fmn), 1}
}

type Animation struct {
	current  F32Vec2
	target   F32Vec2
	lifeTime float32
	finished bool
}

// Lifetime is the life of the animation. Animation speed is depends of the
// delta time and lifetime. For lifeTime parameter, 1.0 value is 1 seconds
func CreateAnimation(from, to F32Vec2, lifeTime float32) Animation {
	return Animation{
		current:  from,
		target:   to,
		lifeTime: lifeTime,
	}
}

// Returns current position of animation as x and y position
// If animation is finished, returned bool value will be true
func (anim *Animation) GetCurrentStep(deltaTime float32) (F32Vec2, bool) {
	if anim.lifeTime > 0 && deltaTime > 0 && !anim.finished {
		anim.current.X += (anim.target.X - anim.current.X) / (anim.lifeTime / deltaTime)
		anim.current.Y += (anim.target.Y - anim.current.Y) / (anim.lifeTime / deltaTime)
		finishedX := math.Abs(float64(anim.target.X-anim.current.X)) < 0.1
		finishedY := math.Abs(float64(anim.target.Y-anim.current.Y)) < 0.1
		anim.finished = finishedX && finishedY
		return anim.current, anim.finished
	}
	return anim.target, true
}

type AtomicBool int32

func (atomicBool *AtomicBool) Set(value bool) {
	var val int32 = 0
	if value == true {
		val = 1
	}
	atomic.StoreInt32((*int32)(atomicBool), val)
}

func (atomicBool *AtomicBool) Get() bool {
	val := atomic.LoadInt32((*int32)(atomicBool))
	return val != 0
}

func (atomicBool *AtomicBool) WaitUntil(val bool) {
	for atomicBool.Get() != val {
		time.Sleep(time.Microsecond)
	}
}

func boolFromInterface(val interface{}) bool {
	switch val.(type) {
	case bool:
		return val == true
	case int, int32, int64, uint, uint32, uint64:
		return val != 0
	default:
		assert_debug(false, "Value type can not be converted to a bool:", val)
		return false
	}
}

func mergeStringArray(arr []string) string {
	str := ""
	for i, arg := range arr {
		str += arg
		if i < len(arr)-1 {
			str += " "
		}
	}
	return str
}