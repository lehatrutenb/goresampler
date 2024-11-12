export last_commit_hash=$(shell git log --format="%H" -n 1)
export last_plot_dir_name=$(shell ls -lt ./plots/ | tail -1 | awk '{print $$NF}')

git_commit:
	@read -p "Enter commit comments:" commit_comments; \
            git commit -m $$commit_comments
	mv plots/latest plots/$$last_commit_hash
	mv plots/$$last_plot_dir_name plots/latest
