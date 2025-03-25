export last_commit_hash=$(shell git log --format="%H" -n 1)
export baseWave1=./base_waves/base1/

# rename plot files to its commit hashes within commit
gitCommit:
	@read -p "Enter commit comments:" commit_comments ; \
        git commit -m "$$commit_comments" ; \
		export last_commit_hash=$(shell git log --format="%H" -n 1) ; \
		cp -r test go_resampler_archive/$$last_commit_hash ; \
		rm -rf prev_test ; \
		mv test prev_test ; \
	-mkdir test
	-mkdir test/plots
	-mkdir test/reports
	-mkdir test/reports_large
	-mkdir test/audio
	-mkdir test/readme_audio

	-mkdir test/reports/rsm_const
	-mkdir test/reports/rsm_spline
	-mkdir test/reports/rsm_fft

	-mkdir test/reports_large/rsm_const
	-mkdir test/reports_large/rsm_spline
	-mkdir test/reports_large/rsm_fft

	-mkdir test/plots/rsm_const
	-mkdir test/plots/rsm_spline
	-mkdir test/plots/rsm_fft

	-mkdir test/audio/rsm_const
	-mkdir test/audio/rsm_spline
	-mkdir test/audio/rsm_fft

gitCommitNotRmTestReports:
	@read -p "Enter commit comments:" commit_comments ; \
        git commit -m "$$commit_comments" ; \
		export last_commit_hash=$(shell git log --format="%H" -n 1) ; \
		cp test go_resampler_archive/$$last_commit_hash ; \


runPlotting:
	python3 ./internal/test_utils/plots.py  -pib=./test/reports_large -pob=./test/plots -p1="rsm_spline" -p2="rsm_const" -p3="rsm_fft" -p4="rsm_batch" -p5="rsm_auto" --workers-amt=10 # it's written here cause running from go code looks dirty

runPlottingSlow:
	python3 ./internal/test_utils/plots.py  -pib=./test/reports_large -pob=./test/plots -p1="rsm_spline" -p2="rsm_const" -p3="rsm_fft" -p4="rsm_batch" -p5="rsm_auto" --workers-amt=1

#if want to process later better to use -json, but I don't think I want to
# care no -a option in first tee to overwrite last testRes
runTest: clearTestDir
	-go test -count=1 -benchmem -v ./... | tee ./test/!testRes
	make runPlotting

runTestSlow: clearTestDir
	-go test -count=1 -benchmem -v ./... | tee ./test/!testRes
	make runPlottingSlow

clearReadmeDir:
	rm -rf test/readme_audio
	mkdir test/readme_audio
	mkdir test/readme_audio/listenable

clearTestDir:
	rm -rf ./test
	mkdir test
	mkdir test/plots
	mkdir test/reports
	mkdir test/reports_large
	mkdir test/audio
	mkdir test/readme_audio
	mkdir test/readme_audio/listenable


	mkdir test/reports/rsm_const
	mkdir test/reports/rsm_spline
	mkdir test/reports/rsm_fft
	mkdir test/reports/rsm_batch
	mkdir test/reports/rsm_auto

	mkdir test/reports_large/rsm_const
	mkdir test/reports_large/rsm_spline
	mkdir test/reports_large/rsm_fft
	mkdir test/reports_large/rsm_batch
	mkdir test/reports_large/rsm_auto

	mkdir test/plots/rsm_const
	mkdir test/plots/rsm_spline
	mkdir test/plots/rsm_fft
	mkdir test/plots/rsm_batch
	mkdir test/plots/rsm_auto

	mkdir test/audio/rsm_const
	mkdir test/audio/rsm_spline
	mkdir test/audio/rsm_fft
	mkdir test/audio/rsm_batch
	mkdir test/audio/rsm_auto

# CALC ONLY 1 CHANNEL IN RESAMPLING TIME
runBenchmark:
	go test -bench=. ./internal/benchmark/benchmark_utils.go ./internal/benchmark/benchmark_test.go ./internal/benchmark/benchmark_batch_test.go  -cpuprofile profile.bat -args minsamplesamt=500000 | tee ./test/!BenchmarkRes
	go tool pprof -ignore="(.*tearDown)|(.*setup)|(.*New)|(.*Merge2Channels)" -relative_percentages  -pdf profile.bat > ./test/profile5e5Samples.pdf
	mv profile.bat ./test/profile.bat

runBenchmarkCustomWave:
	go test -bench=. ./internal/benchmark/benchmark.go  -cpuprofile profile.bat -args minsampledurationins=60 customwave | tee ./test/!BenchmarkRes
	go tool pprof -ignore="(.*tearDown)|(.*setup)|(.*New)|(.*Merge2Channels)" -relative_percentages  -pdf profile.bat > ./test/profile5e5Samples.pdf
	mv profile.bat ./test/profile.bat

runCreateAudioForReadmeTable:
	go test -bench=. ./internal/benchmark/benchmark_utils.go ./internal/benchmark/benchmark_test.go -args minsampledurationins=60
	cp $$baseWave1/base1_8000.wav ./test/readme_audio/25_FFMPEGRsm_8000.mp4
	cp $$baseWave1/base1_16000.wav ./test/readme_audio/26_FFMPEGRsm_16000.mp4

# to gen paste audio urls downloaded to git to internal/benchmark/audio_urls
runReadmeTableGen:
	go run ./cmd/main.go

addBaseWave:
	@read -p "Enter abs path to wave:" path ; \
        ffmpeg -i $$path -ar 8000 ./base_waves/base4/base4_8000.wav ; \
		ffmpeg -i $$path -ar 11000 ./base_waves/base4/base4_11000.wav ; \
		ffmpeg -i $$path -ar 11025 ./base_waves/base4/base4_11025.wav ; \
		ffmpeg -i $$path -ar 16000 ./base_waves/base4/base4_16000.wav ; \
		ffmpeg -i $$path -ar 44000 ./base_waves/base4/base4_44000.wav ; \
		ffmpeg -i $$path -ar 44100 ./base_waves/base4/base4_44100.wav ; \
		ffmpeg -i $$path -ar 48000 ./base_waves/base4/base4_48000.wav

# just run tests from it
preCheckWorkFlow:
	go test -test.short -v ./... -bench=^$ -tags 'NoBenchmarks'

checkWorkflow:
	$(eval include .env)
	act

runDocs:
	xdg-open "http://localhost:8080/"
	$$GOPATH/bin/pkgsite

preCommit: checkWorkflow runDocs
