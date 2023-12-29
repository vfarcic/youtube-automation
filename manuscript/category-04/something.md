
# [[title]] #

# Additional Info:
# - [[additional-info]]

## Intro

Let's say that you would like to have an ephemeral environment.

More specifically, I'll guess that you would like you and people working with you to have an environment with all the tools needed to work on something and that environment to be ephemeral. We should be able to create it when needed and destroy when we don't. That can be, for example, for development purposes.

TODO: Logo: okteto.png, gitpod.png

Now, if I ask you "how to do that", your solution would probably involve containers in some form or another. You might be spinning Docker containers on your laptop or you might be using remote environments like GitHub Codespaces or Okteto or GitPod or any other of the myriad offerings that are all also based on containers. You might be building your own custom solution and that one is likely also based on containers.

How about pipelines like Jenkins, GitHub Actions, GitLab CI, Tekton, Argo Workflows, or whatever you're using to build, test, and do whatever else you're doing with your applications. Those are also ephemeral environments, of sorts. They are created for specific purpose, and destroyed when we're finished doing whatever we're doing. Thos pipelines are probably based on containers as well.

Containers are everywhere, and that's nowhere more true than in ephemeral environments.

But, what if I tell you that the best way to create and destroy ephemeral environments might not be containers but a very special operating system? Heresy! Viktor is loosing his mind. That might be a plausible explanation, but give me a few minutes to show you what I mean before proclaiming me insane.

Here's the background story behind the tool I will explore today.

I started working on a course related to Crossplane. More info about it is coming soon. For now, what matters is that I wanted to make it as accessible to everyone as possible. More importantly, the course should require quite a few tools installed, yet I did not want to put attendees through the trouble of installing everything required for that course.

TODO: Logos: charm-gum.png, git.png, github.png, kind.png, kubectl (there is no logo so put text instead), yq (there is no logo so put text instead), jq.svg, teller.png, aws.png, azure.png, google-cloud.png

From the very start, I knew that everyone attending the course would need `gum`, `git`, GitHub CLI, `kind`, `kubectl`, `yq`, `jq`, `teller`, and AWS, Azure, or Google CLIs depending on which hyperscaler one chooses. That's just the start and as I'm progressing through the course, there will probably be others.

Installing all those can be time consuming so I wanted to streamline the process and, preferably, allow people to have all those with a single command. I could not create a script that installs all those tools since setting it all up would be different on Mac from Windows from Linux. To be more precise, I could create different scripts depending on the operating system, but that seems like a waste of my time and a huge maintenance burden. Moreover, I wanted to do it in a way that people can not only get all those easily, but also remove them when they're done going through the course. It would be pretentious of me to assume that all my recomendations are mandatory as permanent additions to everyone's laptops and removing them can be painful.

TODO: Thumbnail: oosQ3z_9UEM

Naturally, I started working on container images with the idea that people can just spin up containers based on those images and mount local file system into those containers. But that proved to be problematic for quite a few reasons that I wont go into right now. Eventually, I started searching for a different solution. That was around the time I published the video about Dagger. That one.

TODO: Text: @ZiggleFingers (big)

TODO: Logo: nix.png

That's when @ZiggleFingers, one of the people who watched that video, mentioned Nix. That was it. That was the perfect fit for what I needed. I used it in the past, I liked it, I found it very useful, yet, somehow, it dropped from my daily workflow and from my radar. Now it's back. It's just what I needed, and today I want to share with you some of the features it offers.

## Setup

FIXME: * Install `nix` by following the instructions at https://nix.dev/install-nix.

## Ephemeral Shell Environments with Nix

TODO: Logo: oh-my-zsh.png

This time I'm recording from a virgin computer. It's a Mac without anythiong related to development. The only exception is Oh My Zsh. That's the only thing I cannot live without. Whenever I get a new computer, that's always the first thing I install, especially since Zsh is now installed by default in all Macs.

TODO: Logo: nix.png

Besides Oh My ZSH, I also installed Nix, which is what we're exploring today. That's it. Apart from Oh My ZSH and Nix, this computer is virgin. It does not have anythign else. It's a great way to demonstrate Nix by pretending that I am one of the attendees of my course.

The first step is to clone the repo.

TODO: Overlay: screen-01

```sh
gh repo clone vfarcic/crossplane-tutorial
```

It fails on the first step. I don't have `gh`, which is the GitHub CLI. Should I install it? Maybe I should if it turns out I will be using it often. Otherwise, it will be yet another tool I installed but am not using. It would be a waste that I would probably forget to remove.

For now, it would be great if I could get it temporarily, and that's where Nix Shell comes in.

I will start a new Shell, just as I would start Z Shell or Bash but, this time, it will be Nix Shell. Apart from starting the Shell, I'll let it know that new Shell should have `gh`, `kubectl`, and `aws` CLIs.

```sh
nix-shell --packages gh kubectl awscli2
```

That's it. I started a new Shell and, before we continue, I will change the prompt so that it does not occupy valuable realestate space in the video.

Bear in mind this is not the first time I'm running Nix on this computer so it was instant since it used cached packages. In your case, it will probably take a few moments longer since it will need to download those packages the first time.

```sh
PS1="$ "
```

Let me try to clone the repo again. Remember, I do NOT have `gh` on my laptop.

```sh
gh repo clone vfarcic/crossplane-tutorial
```

This time it worked. I cloned the repo since `gh` exists in that Shell session which, by itself, is ephemeral. Once I exit the session, `gh` will be gone, but cached.

As a proof that `gh` exists in Nix alone and not in my "standard" Shell, I can execute `which gh` and we can see that it comes from `/nix/store` which, in a simplified version of the explanation, can be considered Nix cache.

```sh
which gh
```

Similarly, we can see that `kubectl` is available as well and you need to trust me that I did not install it on my Mac.

```sh
which kubectl
```

What is truly great about this way of getting dependencies is that it does not matter whether I am on a Mac, or Windows, or Linux. I don't need to worry whether it should be Brew, or Chocolatey, or APT, or whatever you're using to install packages. Its Nix package model and it works everywhere.

Let me continue pretending that I'm going through the tutorial.

The next step is to enter the directory of the cloned repository...

```sh
cd crossplane-tutorial
```

...make a script executable...

```sh
chmod +x setup/01-managed-resources.sh
```

...and execute it.

```sh
./setup/01-managed-resources.sh
```

TODO: Logos: charm-gum.png, kind.png, kubectl (there is no logo so put text instead), yq (there is no logo so put text instead), aws.png, azure.png, google-cloud.png

That failed as well. That script needs `gum`. Now, if it would be only that one, I could "force" people to install it but that's not the only one. The script creates a Kind cluster, so `kind` CLI is a requirement, besides Docker, we need `kubectl` to interact with that cluster, `yq` to parse YAML, and AWS, Azure, or Google Cloud CLIs depending on which hyperscaler one wants to use.

One option to provide all those can be to extend the `nix-shell` command I run earlier with all those other packages, but there is a better way. I could have created the script itself based on `nix-shell` instead of `bash`. As a matter of fact, I already did that so let's take a look at it.

```sh
cat setup/01-managed-resources-nix.sh
```

This is, essentially, the same script as the one that failed few moments ago, except for the first three lines.

It starts by defining the environment as `nix-shell` interpreter, just as you would typically define it as `bash` or `sh`. Further on, we're telling `nix` itself to interpret it as `bash`. Finally, we're telling it to install packages like `gum`, `kind`, `kubectl`, and so on.

Let's run it, by making the script executable and...

```sh
chmod +x setup/01-managed-resources-nix.sh
```

...executing it.

```sh
./setup/01-managed-resources-nix.sh
```

This time, some packages are not already available in my local cache so it will take a bit of time until Nix downloads them

Let me fast forward to the end and of the package download process and...

...the script worked even though I do not have any of those tools installed on the host, at least not directly.

I won't bother you by showing you what that particular script does. That would be a topic for another video. What matters is that it all works without having to install all the tools required for the script to run.

That's awesome, isn't it.

Let me stop the script and...

```sh
# Press `ctrl+c` to stop
```

...run it again.

```sh
./setup/01-managed-resources-nix.sh
```

The second time the script started running almost immediately since all the required packages are already cached.

```sh
# Press `ctrl+c` to stop
```

Using `nix-shell` inside scripts is an improvement over running command like `nix-shell --packages` followed with the list of all the required packages. It's easier to simply execute a script instead of memorizing a potentially very long command.

However, I don't think that creating Nix-specific scripts is a good idea since that would assume that everyone has Nix. Insted, I prefer writing scripts in a "normal" way with `bash` or `sh` as the Shell so that they can be executed by anyone anywhere, and using Nix as an optional help to get the packages we need. So, I don't want to have `nix-shell` script but I also don't want to have commands like `nix-shell --packages` with a lengthy list of packages. Fortunately, there is a way to solve both issues. There is a way to specify which packages we want without making scripts depend on Nix.

Let me exit the Nix Shell,...

```sh
exit
```

...go back to the `crossplane-tutorial` repo we cloned,...

```sh
cd crossplane-tutorial
```

...and show you a special files called `shell.nix`

```sh
cat shell.nix
```

That file is, by convention, executed automatically whenever we execute `nix-shel`. It contains the list of all the packages Nix Shell should load by default.

It's written in Nix-specific language which, in general, might require time to learn but, fortunately, if all we need is to define the packages, it is pretty straightforwrd and you will probably just copy it, paste it, and change the list of packages.

Over there, I'm specifying that `gum`, `git`, `gh`, and other packages should be installed every time I run `nix-shell` from that directory. As a result, there is no need for me to list all those packages with the `--packages` argument.

I can simply run `nix-shell`...

```sh
nix-shell
```

...wait until dependencies that were not already cached are downloaded and...

...that's it.

Now I'll change the prompt to gain some real-estate and, for example,...

```sh
PS1="$ "
```

...execute `gum` and...

```sh
gum
```

...it works, even though I don't have it on my host nor I had to specify it explicitly.

Moreover, now I can run the "norma" script, the one based on `sh` and not on `nix-shell`. That's the one that was failing before and...

```sh
./setup/01-managed-resources.sh
```

...now it works.

Let me stop the script and...

```sh
# Press `ctrl+c` to stop
```

...exit the Shell,...

```sh
exit
```

...before I show you one more minor but very important thing.

TODO: Logo: oh-my-zsh.png

While I like Nix's capability to deal with packages, I do not particularly like Nix Shell itself. I, for example, prefer using ZSH with Oh My ZSH. It gives me coloring, auto-complete, and other nice features I'm used to.

So, the question is whether we can combine the capability of Nix Shell to manage packages while still running whichever Shell you're used to work in.

The short answer is YES!

We can add `--run` argument and specify which Shell should be executed or simply use the `SHELL` variable to use whichever Shell is currently used which, as I already mentioned, will be ZSH in my case.

```sh
nix-shell --run $SHELL
```

You can see that the prompt is now different from what Nix Shell shows. It's ZSH which I configured to differ depending on whether I'm running it directly from the host or inside another Shell. I won't go into details how I did that. What matters is that now I'm using my favorite Shell inside Nix Shell. I'm combining best of both.

TODO: Miki: Ignore the screen between 09:12 and 09:27.

```sh
exit
```

Now, you might be wondering how I knew what are the names of the packages I choose to use. Some of them are easy to guess like `gum`, `kind`, and `kubectl` since they are named the same as CLIs. Others, like `google-cloud-sdk` instead of `gcloud`, and `awscli2` instead of `aws` cannot be guessed easily.

TODO: Overlay: nixos-search; Lower-third: https://search.nixos.org

Fortunately, search.nixos.org allows us to easily find any package we might need. I can, for example, search for `kubectl`, which happens to have the package with the same name, or for `gcloud`, which happens to have two packages for no good reason, neither of those called `gcloud`. If in doubt, we can click the `Homepage` link to confirm whether the package represents what we really need.

One last thing, for now.

I mentioned that Nix Shell helps us creating ephemeral environments, and that was true, in a way. When we exit the Shell, everything is gone, except the cache itself. Cache keeps accupying disk space and, if needed, we can get rid of it as well with a simple `nix-store --gc` command.

TODO: Overlay: nix-store-gc

```sh
nix-store --gc
```

That's it, when demo is concerned. Now we can talk about Nix. More specifically, let's see what it's good for, whether you should use it, and whatever else comes to my mind.

## Nix Pros and Cons

Nix is much more than Nix Shell. It is, first and foremost, a package managed for Linux. There is a Linux distribution called NixOS, a language built specifically for Nix, and quite a few other things. All those are very interesting and I invite you to check them out.

However, today I focused on Nix as a way to create ephemeral environments. I showed how it manages ephemeral environments on my laptop.

Did it deliver what I need? It does. It sure does.

I feel ashamed that I forgot about it and stopped using it. Now I'm back into it and I will be using it on the material for the upcoming course I'm working on as well as for a few projects I'm working on and, if after a while I do not notice any hidden downside, I will likely use it for all the upcoming projects as well.

Now, even though I used it on my laptop, Nix is not limited to it.

TODO: Logos: jenkins.png, github-actions.png, tekton.png, argo-workflows.png

Another potential use-case for Nix are pipelines. Just as it allows me to bring all the packages I need on my laptop, it can do the same in pipelines like Jenkins, GitHub Actions, Tekton, Argo Workflows, or whatever you're using. However, given the ephemeral nature of pipelines themselves, that might not be a very optimal solution. Cache is likely going to dissapear between execution of pipeline builds. I think that pre-built container images are a better choice for pipelines. On the other hand, some of you do not use pipelines with images based on an OS like Ubuntu in which you download what you need every time a build is executed. In those cases, first of all, you should be ashamed for doing something that silly. But, if you continue doing that, Nix is probably a better choice.

Still... Don't do it. Use pre-built container images.

All in all, Nix Shell is great and I strongly recommend it for ephemeral environments on laptops and desktops. I'm not sure whether that's just as good of a recommendation for pipelines though.

As for everything else Nix offers... It's up to you to explore it, or tell me in the comments whether you'd like me to create a video about the rest of the features of Nix.

## Destroy

```sh
cd ..

rm -rf crossplane-tutorial
```

FIXME: * Delete the fork from [GitHub](https://github.com)
