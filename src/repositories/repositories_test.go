package repositories_test

import (
	"errors"
	"io/fs"
	"log"
	"monitor2/src/repositories"
	"monitor2/src/db/models"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func createFile(file_path string, contents string) {
	if _, err := os.Stat(file_path); errors.Is(err, fs.ErrNotExist) {
		file, err := os.Create(file_path)
		if err != nil {
			log.Fatal(err)
		}

		_, err = file.WriteString(contents)
		if err != nil {
			log.Fatal(err)
		}

		file.Close()
	} else {
    file, err := os.OpenFile(file_path, os.O_WRONLY, 0o644)
		if err != nil {
			log.Fatal(err)
		}

    _, err = file.WriteString(contents)
		if err != nil {
			log.Fatal(err)
		}

    file.Close()
	}
}

func createRepo(path string) *git.Repository {
	repo, err := git.PlainInit(path, false)
	if err != nil {
		log.Fatal(err)
	}
	return repo
}

func createCommit(msg string, repo *git.Repository) {
	_t, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}

	_, err = _t.Add(".")
	if err != nil {
		log.Fatal(err)
	}
	_, err = _t.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Your Name",
			Email: "your.email@example.com",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestGitPullAndDiffMultipleCommits(t *testing.T) {
	path := "/tmp/somerepo"
	os.RemoveAll("/tmp/somerepo")
	os.RemoveAll("/tmp/somerepo2")

	repo := createRepo(path)
	createFile(path+"/README.md", "# README 123\n")
	createCommit("Initial Commit", repo)

	repositories.GitClone("file:///tmp/somerepo", "/tmp/somerepo2")

	createFile(path+"/README2.md", "# README 123\n")
	createCommit("Commit2", repo)

	createFile(path+"/README3.md", "# README 456\n")
	createCommit("Commit3", repo)

	createFile(path+"/blabla.md", "# README 4568\n")
	createCommit("Commit4", repo)

	_repo := models.Repository{
		Directory:     "/tmp/somerepo2",
		WatchedFiles: []byte("[\"README\"]"),
	}

	pull_opts := git.PullOptions{
		RemoteName: "origin",
	}

	diff, _, err := repositories.GitPullAndDiff(_repo, pull_opts)

 	if err != nil {
		t.Fail()
	}
 
  log.Printf("DEBUGPRINT[2]: repositories_test.go:108: diff=%+v\n", diff)
  correct := `a/README2.md b/README2.md
new file mode 100644
index 0000000000000000000000000000000000000000..25dbca072e21969dcc08b588ef9aa1384e131f33
--- /dev/null
+++ b/README2.md
@@ -0,0 +1 @@
+# README 123
a/README3.md b/README3.md
new file mode 100644
index 0000000000000000000000000000000000000000..92c0ec54eb88d7a24a4d605e700365979f3cc486
--- /dev/null
+++ b/README3.md
@@ -0,0 +1 @@
+# README 456
`

	if diff != correct {
		t.Fail()
	}
}

func TestGitPullAndDiff(t *testing.T) {
	path := "/tmp/somerepo"
	os.RemoveAll("/tmp/somerepo")
	os.RemoveAll("/tmp/somerepo2")

	repo := createRepo(path)
	createFile(path+"/README.md", "Hello, world!\n123\n321\n")
	createFile(path+"/README1.md", "README 2 @\n")
	createCommit("Initial commit", repo)

	repositories.GitClone("file:///tmp/somerepo", "/tmp/somerepo2")

	createFile(path+"/README.md", "Hello, world!\n321\n123\n")
	createFile(path+"/README1.md", "README 2 @ NEW STRING 123\n")
	createCommit("Second commit", repo)

	_repo := models.Repository{
		Directory:     "/tmp/somerepo2",
		WatchedFiles: []byte("[\"README.md\", \"README1.md\"]"),
	}

	pull_opts := git.PullOptions{
		RemoteName: "origin",
	}

	diff, _, err := repositories.GitPullAndDiff(_repo, pull_opts)
	if err != nil {
		t.Fail()
	}

	correct := `a/README.md b/README.md
index 5239431b434f2d2dcb241d725f4569cdb46d3f05..4713961164e91c926bbfc9c89d0c1044b644ce4c 100644
--- a/README.md
+++ b/README.md
@@ -1,3 +1,3 @@
 Hello, world!
-123
 321
+123
a/README1.md b/README1.md
index 858c753f71f2bdd7417898c0e301d394e33e91b7..d9a4214099c8ecd17b5d84b688aa0ccbf75cfd5c 100644
--- a/README1.md
+++ b/README1.md
@@ -1 +1 @@
-README 2 @
+README 2 @ NEW STRING 123
`

	if diff != correct {
		t.Fail()
	}
}

func TestParseDiff(t *testing.T) {
	diff := `diff --git a/test/README.md b/test/README.md
index 5239431b434f2d2dcb241d725f4569cdb46d3f05..ccb9dcb7fab6a1c94a32cfa70f2d504f552311d7 100644
--- a/test/README.md
+++ b/test/README.md
@@ -1,3 +1,6 @@
Hello, world!
123
321
+Hello, world!
+321
+123
diff --git a/test/README1.md b/test/README1.md
index 858c753f71f2bdd7417898c0e301d394e33e91b7..e563f0861e50ef892f559836075414abff98f2b3 100644
--- a/test/README1.md
+++ b/test/README1.md
@@ -1 +1,2 @@
README 2 @
+README 2 @ NEW STRING 123
`

	res := repositories.ParseDiff(diff, []string{"README.md"})
	correct := `a/test/README.md b/test/README.md
index 5239431b434f2d2dcb241d725f4569cdb46d3f05..ccb9dcb7fab6a1c94a32cfa70f2d504f552311d7 100644
--- a/test/README.md
+++ b/test/README.md
@@ -1,3 +1,6 @@
Hello, world!
123
321
+Hello, world!
+321
+123
`

	if res != correct {
		t.Fail()
	}
}

func TestParseDiffDoesntAddTwice(t *testing.T) {
	diff := `a/test/README.md b/test/README.md
index 5239431b434f2d2dcb241d725f4569cdb46d3f05..ccb9dcb7fab6a1c94a32cfa70f2d504f552311d7 100644
--- a/test/README.md
+++ b/test/README.md
@@ -1,3 +1,6 @@
Hello, world!
123
321
+Hello, world!
+321
+123
a/test/README.md b/test/README.md
index 5239431b434f2d2dcb241d725f4569cdb46d3f05..ccb9dcb7fab6a1c94a32cfa70f2d504f552311d7 100644
--- a/test/README.md
+++ b/test/README.md
@@ -1,3 +1,6 @@
Hello, world!
123
321
+Hello, world!
+321
+123`

	res := repositories.ParseDiff(diff, []string{"README.md", "README.md"})
	if len(res) > len(diff) {
		t.Fail()
	}
}

func TestParseDiffIgnoresOnBody(t *testing.T) {
	diff := `a/test/README.md b/test/README.md
index 5239431b434f2d2dcb241d725f4569cdb46d3f05..ccb9dcb7fab6a1c94a32cfa70f2d504f552311d7 100644
--- a/test/README.md
+++ b/test/README.md
@@ -1,3 +1,6 @@
Hello, world! diff --git 1
123
321
+Hello, world!
+321
+123`

	res := repositories.ParseDiff(diff, []string{"README.md"})
	if res != diff {
		t.Fail()
	}
}
