export last_commit_hash=$(shell git log --format="%H" -n 1)
export last_plot_dir_name=$(shell ls -lt ./plots/ | tail -1 | awk '{print $$NF}')

# rename plot files to its commit hashes within commit
git_commit:	
	@read -p "Enter commit comments:" commit_comments ; \
        git commit -m "$$commit_comments" ; \
		export last_commit_hash=$(shell git log --format="%H" -n 1) ; \
		mv plots/$$last_plot_dir_name plots/$$last_commit_hash
	-mkdir plots/latest