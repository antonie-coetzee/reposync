package main

import (
	"gopkg.in/src-d/go-git.v4"
	"os"
	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
	log "github.com/Sirupsen/logrus"
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
		log.Error(err)
	}
}

func handlePush(payload interface{}, header webhooks.Header) {	
	log.Info("push event received")
	log.Infof("filtering on branch: '%s'", branch)	
	pl := payload.(github.PushPayload)		
	refBranch := "refs/heads/" +  branch 		
	if pl.Ref == refBranch {
		log.Infof("pushed to '%s', processing", refBranch)
		defer func() { 
			if r := recover(); r != nil {
				log.Error(r)
			}
		}()
		log.Infof("checking if working directory exists at: %s", workingDir)
		if val, _ := exists(workingDir); val {
			log.Info("working directory exists")
			r, err := git.PlainOpen(workingDir)
			if err != nil && err.Error() == "repository not exists" {
				log.Error(err)
				log.Warn("cleaning working directory")
				os.RemoveAll(workingDir)
				clone()
				return
			}
			log.Info("valid repository detected")
			w, _ := r.Worktree()
			log.Info("hard resetting working directory")
			w.Reset(&git.ResetOptions{Mode: git.HardReset})
			log.Info("pulling latest changes")
			err = w.Pull(&git.PullOptions{RemoteName: "origin"})
			if err != nil && err.Error() == "already up-to-date"{
				log.Info("already up-to-date")
				return
			}
			log.Info("working directory now up-to-date")
		}else{
			clone()
		}
	}else{
		log.Infof("pushed to '%s', not processing", pl.Ref)
	}
}

func clone() {
	log.Info("cloning repository")
	git.PlainClone(workingDir, false, &git.CloneOptions{
		URL:repoURL,
		Progress: os.Stdout,
	})
	log.Info("working directory now up-to-date")	
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

