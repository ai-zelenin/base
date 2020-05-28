package util

import (
	"math"
)

// MinMax return min and max value
func MinMax(array []float64) (float64, float64) {
	if len(array) > 0 {
		var max = array[0]
		var min = array[0]
		for _, value := range array {
			if max < value {
				max = value
			}
			if min > value {
				min = value
			}
		}
		return min, max
	}
	return 0, 0
}

// ArithmeticMean calculate mean from sequence of numbers
func ArithmeticMean(array []float64) float64 {
	if len(array) > 0 {
		n := float64(len(array))
		m := 0.0
		for _, value := range array {
			m += value
		}
		return m / n
	}
	return 0
}

// ArithmeticMeanWeights calculate v/am array for all v in array
func ArithmeticMeanWeights(values []float64) []float64 {
	am := ArithmeticMean(values)
	weights := make([]float64, len(values))
	for i, v := range values {
		weights[i] = am / v
	}
	return weights
}

// WeighedArithmeticMean calculate v/am array for all v in array
func WeighedArithmeticMean(values []float64, weights []float64) float64 {
	diff := len(values) - len(weights)
	if diff > 0 {
		for i := 0; i < diff; i++ {
			weights = append(weights, 0)
		}
	}
	var valuesSum float64
	var weightSum float64
	for i := range values {
		valuesSum += weights[i] * values[i]
		weightSum += weights[i]
	}
	return valuesSum / weightSum
}

// StandardDeviation from sequence of numbers
func StandardDeviation(array []float64) float64 {
	if len(array) > 0 {
		n := float64(len(array))
		am := ArithmeticMean(array)
		ds := 0.0
		for _, a := range array {
			da := math.Pow(am-a, 2)
			ds += da
		}
		return math.Pow(ds/n, 0.5)
	}
	return 0
}

// GeneralizedMean calculate generalized mean with d = factor
func GeneralizedMean(values []float64, factor float64) float64 {
	if len(values) > 0 {
		Ev := 0.0
		n := float64(len(values))
		for _, v := range values {
			Ev += math.Pow(v, factor)
		}
		return math.Pow(Ev/n, 1/factor)
	}
	return 0
}

// GeometricMean calculate mean from sequence of numbers
func GeometricMean(array []float64) float64 {
	if len(array) > 0 {
		n := float64(len(array))
		m := 1.0
		for _, value := range array {
			m *= value
		}
		return math.Pow(m, 1.0/n)
	}
	return 0
}

// WeightedGeometricMean calculate mean from sequence of numbers and its weights
func WeightedGeometricMean(valuesArray []float64, weightArray []float64) float64 {
	if len(valuesArray) > 0 && len(weightArray) == len(valuesArray) {
		d := len(valuesArray) - len(weightArray)
		if d > 0 {
			for ; d == 0; d-- {
				weightArray = append(weightArray, 1)
			}
		}
		Ew := 0.0
		Pv := 1.0
		for _, w := range weightArray {
			Ew += w
		}
		for vIndex, v := range valuesArray {
			Pv *= math.Pow(v, weightArray[vIndex])
		}
		return math.Pow(Pv, 1.0/Ew)
	}
	return 0
}

// RoundToDigit rounds number to N digits after dot
func RoundToDigit(v, n float64) float64 {
	var m = math.Pow(10, n)
	return math.RoundToEven(v*m) / m
}
