package webvtt

var SampleVTT = []byte(`WEBVTT

00:00:00.500 --> 00:00:02.000
The Web is always changing

00:00:02.500 --> 00:00:04.300
and the way we access it is changing

00:00:04.500 --> 00:00:05.300
and the way we access it is changing
`)

var InvalidVTT1 = []byte(`WEBVT`)
var InvalidVTT2 = []byte(`hello world`)
