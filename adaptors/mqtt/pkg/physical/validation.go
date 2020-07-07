package physical

import (
	"github.com/pkg/errors"
)

// verifyWritableJSONPath verifies the legality of path which is used for writing value,
// this inspires by https://github.com/tidwall/sjson#path-syntax, in order to further ensure the ability to write back,
// the following checks were made:
// - `children.-1` is not allowed;
// - `children|@reverse` is not allowed;
// - `child*.2` is not allowed;
// - `c?ildren.0` is not allowed;
// - `friends.#.first` is not allowed.
func verifyWritableJSONPath(path string) error {
	for i := 0; i < len(path); i++ {
		var p = path[i]
		switch p {
		case '\\':
			i++
			if i < len(path) {
				i++
			}
		case '.':
			var next = path[i+1:]
			if len(next) > 1 && next[0] == '-' {
				return errors.New("minus character not allowed in path")
			}
		case '*', '?':
			return errors.New("wildcard characters not allowed in path")
		case '#':
			return errors.New("array access character not allowed in path")
		case '@':
			return errors.New("modifiers not allowed in path")
		case '|':
			return errors.New("pipe characters not allowed in path")
		}
	}
	return nil
}
