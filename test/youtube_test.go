package test

import (
	"discord-music-bot/services"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDuration(t *testing.T) {
	assert := assert.New(t)

	duration := "P1DT23H45M20S"
	offset := "25877s"
	expected := "1:16:34:03"
	actual := services.getDuration(duration, offset)
	assert.Equal(expected, actual, "the data have to be equal")

	duration = "P1DT"
	offset = "5000s"
	expected = "22:36:40"
	actual = services.getDuration(duration, offset)
	assert.Equal(expected, actual, "the data have to be equal")

	duration = "PT1H"
	offset = "300s"
	expected = "55:00"
	actual = services.getDuration(duration, offset)
	assert.Equal(expected, actual, "the data have to be equal")

	duration = "PT1M"
	offset = "20s"
	expected = "00:40"
	actual = services.getDuration(duration, offset)
	assert.Equal(expected, actual, "the data have to be equal")

	duration = "PT4H2S"
	offset = "260s"
	expected = "03:55:42"
	actual = services.getDuration(duration, offset)
	assert.Equal(expected, actual, "the data have to be equal")

}
