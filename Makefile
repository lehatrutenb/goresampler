export last_commit_hash=$(shell git log --format="%H" -n 1)
export last_plot_dir_name=$(shell ls -lt ./plots/ | head -2 | tail -1 | awk '{print $$NF}')
#export last_plot_dir_name=latest

# rename plot files to its commit hashes within commit
gitCommit:
	@read -p "Enter commit comments:" commit_comments ; \
        git commit -m "$$commit_comments" ; \
		export last_commit_hash=$(shell git log --format="%H" -n 1) ; \
		mv plots/$$last_plot_dir_name plots/$$last_commit_hash
	-mkdir plots/latest
	-mkdir plots/latest/legacy

runPlotting:
	python3 ./resampler/internal/test_utils/plots.py --plot-path1="./plots/latest/" --plot-path2="./plots/latest/legacy/" --workers-amt=20 # it's written here cause running from go code looks dirty


#if want to process later better to use -json, but I don't think I want to
runTest:
	-go test -count=1 -bench=. -benchmem -v ./resampler/internal/resample/ | tee ./plots/latest/!testRes
	-go test -count=1 -bench=. -benchmem -v ./resampler/internal/resample/resamplerl | tee ./plots/latest/legacy/!testRes
	make runPlotting

clearTestDir:
	rm -rf  ./plots/latest
	mkdir plots/latest
	mkdir plots/latest/legacy

runGenerate:
	-rm resampler/internal/resample/resamplerl/legacy_resample_test.go
	go run ./resampler/internal/test_utils/legacy_gen/gen_legacy_tests.go -o ./resampler/internal/resample/resamplerl/legacy_resample_test.go