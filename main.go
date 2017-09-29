package main

import (
	"gopkg.in/src-d/go-git.v4"
	"os"
	"fmt"
	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

func getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}

var (
	hookPath = getenv("hookPath","/")
	repoURL = getenv("repoUrl","https://github.com/WireJunky/BlogContent.git")
	workingDir = getenv("workingDir","/tmp/working")
	branch = getenv("branch","master")
	port = getenv("port","8081")
	secret = getenv("secret","")
)

func main() {
	hook := github.New(&github.Config{Secret: secret})
	hook.RegisterEvents(handlePush, github.PushEvent)

	err := webhooks.Run(hook, ":"+port, hookPath)
	if err != nil {
		fmt.Println(err)
	}
}

func handlePush(payload interface{}, header webhooks.Header) {	
	fmt.Println("push event received")
	fmt.Printf("filtering on branch: '%s'\n", branch)	
	pl := payload.(github.PushPayload)		
	refBranch := "refs/heads/" +  branch 		
	if pl.Ref == refBranch {
		fmt.Printf("pushed to '%s', processing\n", refBranch)
		defer func() { 
			if r := recover(); r != nil {
				fmt.Println("error:", r)
			}
		}()

		if val, _ := exists(workingDir); val {
			fmt.Println("pulling latest changes")
			r, _ := git.PlainOpen(workingDir)
			w, _ := r.Worktree()
			err := w.Pull(&git.PullOptions{RemoteName: "origin"})
			if err != nil && err.Error() == "already up-to-date"{
				fmt.Println("already up-to-date")
				return
			}
			fmt.Println("working directory now up-to-date")
		}else{
			fmt.Println("cloning repository")
			git.PlainClone(workingDir, false, &git.CloneOptions{
				URL:repoURL,
				Progress: os.Stdout,
			})
			fmt.Println("working directory now up-to-date")
		}
	}else{
		fmt.Printf("pushed to '%s', not processing\n", pl.Ref)
	}
}


func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

