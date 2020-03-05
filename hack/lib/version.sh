#!/usr/bin/env bash

##
# Borrowed from github.com/kubernetes/kubernetes/hack/lib/version.sh
##

# -----------------------------------------------------------------------------
# Version management helpers. These functions help to set the
# following variables:
#
#        OCTOPUS_GIT_COMMIT  -  The git commit id corresponding to this
#                               source code.
#    OCTOPUS_GIT_TREE_STATE  -  "clean" indicates no changes since the git commit id
#                               "dirty" indicates source code changes after the git commit id
#                               "archive" indicates the tree was produced by 'git archive'
#       OCTOPUS_GIT_VERSION  -  "vX.Y" used to indicate the last release version.
#         OCTOPUS_GIT_MAJOR  -  The major part of the version
#         OCTOPUS_GIT_MINOR  -  The minor component of the version
#        OCTOPUS_BUILD_DATE  -  The build date of the version

function octopus::version::get_version_vars() {
  OCTOPUS_BUILD_DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

  # if the source was exported through git archive, then
  # use this to get the git tree state
  # shellcheck disable=SC2016,SC2050
  if [[ '$Format:%%$' == "%" ]]; then
    OCTOPUS_GIT_COMMIT='$Format:%H$'
    OCTOPUS_GIT_TREE_STATE="archive"
    # when a 'git archive' is exported, the '$Format:%D$' below will look
    # something like 'HEAD -> release-1.8, tag: v1.8.3' where then 'tag: '
    # can be extracted from it.
    if [[ '$Format:%D$' =~ tag:\ (v[^ ,]+) ]]; then
      OCTOPUS_GIT_VERSION="${BASH_REMATCH[1]}"
    fi
  fi

  local git=(git --work-tree "${ROOT_DIR}")

  if [[ -n ${OCTOPUS_GIT_COMMIT-} ]] || OCTOPUS_GIT_COMMIT=$(git rev-parse "HEAD^{commit}" 2>/dev/null); then
    if [[ -z ${OCTOPUS_GIT_TREE_STATE-} ]]; then
      # check if the tree is dirty, default is dirty.
      if git_status=$(git status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
        OCTOPUS_GIT_TREE_STATE="clean"
      else
        OCTOPUS_GIT_TREE_STATE="dirty"
      fi
    fi

    # use git describe to find the version based on tags,
    # this translates the "git describe" to an actual semver.org,
    # compatible semantic version that looks something like this:
    #   v1.1.0-alpha.0.6+84c76d1142ea4d
    if [[ -n ${OCTOPUS_GIT_VERSION-} ]] || OCTOPUS_GIT_VERSION=$(git describe --tags --abbrev=14 "${OCTOPUS_GIT_COMMIT}^{commit}" 2>/dev/null); then
      # shellcheck disable=SC2001
      DASHES_IN_VERSION=$(echo "${OCTOPUS_GIT_VERSION}" | sed "s/[^-]//g")
      if [[ "${DASHES_IN_VERSION}" == "---" ]]; then
        # distance to subversion (v1.1.0-subversion-1-gCommitHash)
        # shellcheck disable=SC2001
        OCTOPUS_GIT_VERSION=$(echo "${OCTOPUS_GIT_VERSION}" | sed "s/-\([0-9]\{1,\}\)-g\([0-9a-f]\{14\}\)$/.\1\+\2/")
      elif [[ "${DASHES_IN_VERSION}" == "--" ]]; then
        # distance to base tag (v1.1.0-1-gCommitHash)
        # shellcheck disable=SC2001
        OCTOPUS_GIT_VERSION=$(echo "${OCTOPUS_GIT_VERSION}" | sed "s/-g\([0-9a-f]\{14\}\)$/+\1/")
      fi
      if [[ "${OCTOPUS_GIT_TREE_STATE}" == "dirty" ]]; then
        # git describe --dirty only considers changes to existing files, but
        # that is problematic since new untracked .go files affect the build,
        # so use git status instead.
        OCTOPUS_GIT_VERSION+="-dirty"
      fi

      # try to match the "git describe" output to a regex to try to extract
      # the "major" and "minor" versions and whether this is the exact tagged
      # version or whether the tree is between two tagged versions.
      if [[ "${OCTOPUS_GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?([-].*)?([+].*)?$ ]]; then
        OCTOPUS_GIT_MAJOR=${BASH_REMATCH[1]}
        OCTOPUS_GIT_MINOR=${BASH_REMATCH[2]}
        if [[ -n "${BASH_REMATCH[4]}" ]]; then
          OCTOPUS_GIT_MINOR+="+"
        fi
      fi

      # if OCTOPUS_GIT_VERSION is not a valid Semantic Version, then refuse to build.
      if ! [[ "${OCTOPUS_GIT_VERSION}" =~ ^v([0-9]+)\.([0-9]+)(\.[0-9]+)?(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then
        octopus::log::error "OCTOPUS_GIT_VERSION should be a valid Semantic Version.
        Current value is: ${OCTOPUS_GIT_VERSION}
        Please see more details here: https://semver.org"
      fi
    else
      OCTOPUS_GIT_VERSION=$(git rev-parse --abbrev-ref HEAD | sed -E 's/[^a-zA-Z0-9]+/-/g')
      if [[ "${OCTOPUS_GIT_TREE_STATE}" == "dirty" ]]; then
        # git describe --dirty only considers changes to existing files, but
        # that is problematic since new untracked .go files affect the build,
        # so use git status instead.
        OCTOPUS_GIT_VERSION+="-dirty"
      fi
    fi
  fi
}


