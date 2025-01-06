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

runTest:
	go test -count=1 -v ./resampler/internal/resample/
	python3 ./resampler/internal/test_utils/plots.py --plot-path="./plots/latest/" # it's written here cause running from go code looks dirty
