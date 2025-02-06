export last_commit_hash=$(shell git log --format="%H" -n 1)
#export last_plot_dir_name=$(shell ls -lt ./test/ | head -2 | tail -1 | awk '{print $$NF}')
#export last_plot_dir_name=latest

# rename plot files to its commit hashes within commit
gitCommit:
	@read -p "Enter commit comments:" commit_comments ; \
        git commit -m "$$commit_comments" ; \
		export last_commit_hash=$(shell git log --format="%H" -n 1) ; \
		mv test go_resampler_archive/$$last_commit_hash
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


runPlotting:
	python3 ./internal/test_utils/plots.py  -pib=./test/reports_large -pob=./test/plots -p1="rsm_spline" -p2="rsm_const" -p3="rsm_fft" --workers-amt=20 # it's written here cause running from go code looks dirty


#if want to process later better to use -json, but I don't think I want to
# care no -a option in first tee to overwrite last testRes
runTest:
	-go test -count=1 -bench=. -benchmem -v ./internal/resample/resamplerfft | tee ./test/!testRes
	-go test -count=1 -bench=. -benchmem -v ./internal/resample/resamplerspline | tee -a ./test/!testRes
	-go test -count=1 -bench=. -benchmem -v ./internal/resample/resamplerce | tee -a ./test/!testRes
	make runPlotting

clearReadmeDir:
	rm -rf ./cmd/out
	mkdir cmd/out

clearTestDir:
	rm -rf ./test
	mkdir test
	mkdir test/plots
	mkdir test/reports
	mkdir test/reports_large
	mkdir test/audio
	mkdir test/readme_audio


	mkdir test/reports/rsm_const
	mkdir test/reports/rsm_spline
	mkdir test/reports/rsm_fft

	mkdir test/reports_large/rsm_const
	mkdir test/reports_large/rsm_spline
	mkdir test/reports_large/rsm_fft

	mkdir test/plots/rsm_const
	mkdir test/plots/rsm_spline
	mkdir test/plots/rsm_fft

	mkdir test/audio/rsm_const
	mkdir test/audio/rsm_spline
	mkdir test/audio/rsm_fft

#runGenerate:
#-rm resampler/internal/resample/resamplerl/legacy_resample_test.go
#go run ./resampler/internal/test_utils/legacy_gen/gen_legacy_tests.go -o ./resampler/internal/resample/resamplerl/legacy_resample_test.go