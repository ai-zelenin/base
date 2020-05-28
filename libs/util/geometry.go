package util

import (
	"math"
)

type Point struct {
	X, Y float64
}

type Line struct {
	LineFunc LineFunc
}

type LineSegment struct {
	Line
	Start, End Point
}

type LineFunc func(x float64) float64

func FindLineFuncBySegment(line *LineSegment) (lineK float64, lineB float64, lineFunc LineFunc) {
	lineK = (line.Start.Y - line.End.Y) / (line.Start.X - line.End.X)
	lineB = line.End.Y - lineK*line.End.X
	lineFunc = func(x float64) float64 {
		return lineK*x + lineB
	}
	return lineK, lineB, lineFunc
}

func IsLineSegmentsCross(a, b *LineSegment) bool {
	return !(math.Max(a.Start.Y, a.End.Y) < math.Min(b.Start.Y, b.End.Y) || math.Min(a.Start.Y, a.End.Y) > math.Max(b.Start.Y, b.End.Y))
}

func FindCrossPointOfLineSegments(a, b *LineSegment, accuracy float64) (bool, *Point) {
	if a.Start.X < b.Start.X {
		return false, nil
	}
	var aIsVertical = a.Start.X-a.End.X == 0
	var bIsVertical = b.Start.X-b.End.X == 0
	if aIsVertical && bIsVertical {
		return false, nil
	}

	var ok = IsLineSegmentsCross(a, b)
	if !ok {
		return ok, nil
	}

	var crossPoint = new(Point)

	var aK, aB, aF = FindLineFuncBySegment(a)
	var bK, bB, bF = FindLineFuncBySegment(b)
	var kDelta = aK - bK
	if kDelta == 0 {
		return false, nil
	}
	var bDelta = bB - aB
	crossPoint.X = bDelta / kDelta
	var aY = RoundToDigit(aF(crossPoint.X), accuracy)
	var bY = RoundToDigit(bF(crossPoint.X), accuracy)
	if aY == bY {
		crossPoint.Y = aY
	}

	return true, crossPoint
}
