export last_commit_hash=$(shell git log --format="%H" -n 1)
export last_plot_dir_name=$(shell ls -lt ./plots/ | tail -1 | awk '{print $$NF}')

git_push
	mv plots/$$last_plot_dir_name plots/$$last_commit_hash
	git push
