package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	. "github.com/dave/jennifer/jen"
)

func main() {
	outF := flag.String("o", "./resampler/internal/resample/resamplerl/legacy_resample_test.go", "path to output file")
	flag.Parse()

	f := NewFile("resamplerl_test")

	for _, preFrInt := range []int{8000, 11000, 16000, 44100, 48000} {
		for _, preToInt := range []int{8000, 16000} {
			frInt := preFrInt - preFrInt%1000
			toInt := preToInt - preToInt%1000
			if toInt == frInt {
				continue
			}
			frStr := strconv.Itoa(frInt)
			toStr := strconv.Itoa(toInt)
			fr := strconv.Itoa(frInt / 1000)
			to := strconv.Itoa(toInt / 1000)
			leg := "L" // "L" if legacy, "" otherwise
			sName := fmt.Sprintf("resampler%sTo%s%s", fr, to, leg)
			strRepr := fmt.Sprintf("%s_to_%s_resampler%s", frStr, toStr, leg)
			rFuncName := fmt.Sprintf("Resample%sTo%s%s", fr, to, leg)
			tFuncName := fmt.Sprintf("TestResample%sTo%s%s", fr, to, leg)
			sinWavewTime := 55 // sec // 55 to div by 11
			testRunAmt := 10   // amt times to run test - the more the better time measure will get (but > 10 questionable because of caches)

			tUtilsS := "resampler/internal/test_utils"
			assertS := "github.com/stretchr/testify/assert"

			f.Type().Id(sName).Struct(
				Id("resampled").Index().Int16(),
			)

			f.Func().Params(Null().Id(sName)).Id("Copy").Params().Qual(tUtilsS, "TestResampler").Block(
				Return(Op("new").Call(Id(sName))),
			)

			f.Func().Params(Null().Id(sName)).Id("String").Params().String().Block(
				Return(Lit(strRepr)),
			)

			f.Func().Params(Id("rsm").Op("*").Id(sName)).Id("Resample").Params(Id("inp").Index().Int16()).Error().Block(
				Return(Qual("resampler/internal/resample/resamplerl", rFuncName).Call(Id("inp"), Op("&").Id("rsm").Dot("resampled"))),
			)

			f.Func().Params(Id("rsm").Id(sName)).Id("OutLen").Params().Int().Block(
				Return(Id("len").Call(Id("rsm").Dot("resampled"))),
			)

			f.Func().Params(Null().Id(sName)).Id("OutRate").Params().Int().Block(
				Return(Lit(toInt)),
			)

			f.Func().Params(Id("rsm").Id(sName)).Id("Get").Params(Id("ind").Int()).Call(List(Int16(), Error())).Block(
				If(Null(), Id("ind").Op(">=").Op("len").Call(Id("rsm").Dot("resampled"))).Block(
					Return(Lit(0), Qual("errors", "New").Call(Lit("out of bounds"))),
				),
				Return(Id("rsm").Dot("resampled").Index(Id("ind")), Nil()),
			)

			// main test func
			f.Func().Id(tFuncName).Params(Id("t").Op("*").Qual("testing", "T")).Null().Block(
				Defer().Func().Params().Block(
					If(Id("r").Op(":=").Recover(), Id("r").Op("!=").Nil()).Block(
						Id("t").Dot("Error").Call(Id("r")),
					),
				).Call(),

				Var().Id("tObj").Qual(tUtilsS, "TestObj").Op("=").Qual(tUtilsS, "TestObj").Values().Dot("New").Call(
					Qual(tUtilsS, "SinWave").Values().Dot("New").Call(Lit(0), Lit(sinWavewTime), Lit(frInt), Lit(toInt)),
					Qual(tUtilsS, "TestResampler").Call(Op("&").Id(sName).Values()),
					Lit(testRunAmt),
					Id("t"),
					Op("&").Qual(tUtilsS, "TestOpts").Values(Lit(false), Lit("../../../../plots")),
				),

				Id("err").Op(":=").Id("tObj").Dot("Run").Call(),
				If(Null(), Op("!").Qual(assertS, "NoError").Call(Id("t"), Id("err"), Lit("failed to run resampler"))).Block(
					Id("t").Dot("Error").Call(Id("err")),
				),
				Id("err").Op("=").Id("tObj").Dot("Save").Call(Lit("latest/legacy")),
				If(Null(), Op("!").Qual(assertS, "NoError").Call(Id("t"), Id("err"), Lit("failed to save test results"))).Block(
					Id("t").Dot("Error").Call(Id("err")),
				),
			)
		}
	}
	f.HeaderComment("Code generated by gen_legacy_tests.go; DO NOT EDIT.")
	err := f.Save(*outF)
	if err != nil {
		log.Fatal("failed to save generated file", err)
	}
}
