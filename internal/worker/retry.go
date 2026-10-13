package worker


import (
	"time"
	"math/rand/v2"
)

func CalculateBackoff(attempt int) time.Duration{

	var base time.Duration

	switch attempt {
	case 1:
		base = 5 * time.Second
	case 2:
		base = 30 * time.Second
	case 3:
		base = 5 * time.Minute	
	default:
		base = 30 * time.Minute
	}
		
	jitter := time.Duration(rand.Float64() * float64(time.Second))
	return base + jitter
}


func ShouldRetry(statusCode int, err error) bool {

	if err != nil{
		return true
	}

	if statusCode == 408 || statusCode == 429 {
		return true
	}

	if statusCode >= 500 && statusCode <= 504 {
		return true
	}


	return false
}
