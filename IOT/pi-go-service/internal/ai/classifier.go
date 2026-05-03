package ai

import (
	"math"
	"math/rand"
	"time"
)

type Classifier struct {
	thresholds map[string]FrequencyRange
}

type FrequencyRange struct {
	MinHz float64
	MaxHz float64
}

func NewClassifier() *Classifier {
	return &Classifier{
		thresholds: map[string]FrequencyRange{
			"LOCUST": {MinHz: 2000, MaxHz: 3000},
			"RATS":   {MinHz: 15000, MaxHz: 25000},
			"BIRDS":  {MinHz: 1000, MaxHz: 5000},
		},
	}
}

func (c *Classifier) Classify(dominantFreq float64, amplitude float64) string {
	if amplitude < 0.01 {
		return "UNKNOWN"
	}
	for pest, fr := range c.thresholds {
		if dominantFreq >= fr.MinHz && dominantFreq <= fr.MaxHz {
			return pest
		}
	}
	return "UNKNOWN"
}

func (c *Classifier) RandomFreqForPest(pestType string) float64 {
	switch pestType {
	case "rat":
		return 15000 + rand.Float64()*10000
	case "bird":
		return 1000 + rand.Float64()*4000
	case "locust":
		return 2000 + rand.Float64()*1000
	default:
		return 0
	}
}

func RandomDuration() int {
	rand.Seed(time.Now().UnixNano())
	return 2 + rand.Intn(4)
}

func calculateDominantFrequency(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += s
	}
	return sum / float64(len(samples))
}

func calculateAmplitude(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sumSq float64
	for _, s := range samples {
		sumSq += s * s
	}
	return math.Sqrt(sumSq / float64(len(samples)))
}

func AnalyzeSamples(samples []float64) (string, float64, float64) {
	if len(samples) == 0 {
		return "UNKNOWN", 0, 0
	}
	dominantFreq := calculateDominantFrequency(samples)
	amplitude := calculateAmplitude(samples)
	classifier := NewClassifier()
	pestType := classifier.Classify(dominantFreq, amplitude)
	return pestType, dominantFreq, amplitude
}
