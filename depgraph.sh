# the directory where the main function is
# e.g. `.` or `./cmd/app-name`
MAIN_DIR=$1

# the module address of the project, written in `go.mod` file.
# e.g. `github.com/mdaliyan/app-name`
REPO=$(go list -m) # github.com/mdaliyan/app-name

# The address of the project owner, written in `go.mod` file.
# e.g. `github.com/mdaliyan`
REOP_OWNER=$(go list -m | grep -oE '^[^/]+/[^/]+')

# Collect all the direct and indirect depencencies
# The command itself prints each of them in a new line
# but we will get them space separated in one line in
# our environment variable.
PACKAGES_GRAPH="$(go mod graph | sed 's/ /\n/g' | grep -oE '^[^@]+')"
PACKAGES_DEPS="$(go list -deps $MAIN_DIR)"

# Concat all depencencies with a space in between.
# This list contains many duplications.
ALL_DEPENDENCIES="${PACKAGES_GRAPH}${PACKAGES_DEPS:+ }${PACKAGES_DEPS}"

# This command contains multiple pipes:
#  1. replacing space with new line
#  2. keeping the repository owner's addresses in each line
#  3. get a unique list out of them
#  4. remove your own repo owner address from the list
#  ... the rest is to replace new lines with commas and
#      remove the last comma from the end
EXCLUDED_PACKAGES=$(echo $ALL_DEPENDENCIES |
	sed 's/ /\n/g' |
	grep -oE '^[^/]+/[^/]+' |
	uniq |
	grep -v $REOP_OWNER |
	tr '\n' ' ' | sed 's/ $//' | sed 's/ /,/g' | tr '\n' ' ')

# This command contains multiple pipes:
#  1. generates a graph's dot file
#  2. remove your repository owner address from the links
#     from the dot fileso the nodes in the charts become
#     short and nicer to read
#  4. generate the graph in png format
godepgraph -s -ignoreprefixes "$EXCLUDED_PACKAGES" $MAIN_DIR |
	sed "s|\"$MODULE|\"|g" |
	dot -Tpng -o godepgraph.png
