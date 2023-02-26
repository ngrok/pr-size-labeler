# PR Size Labeler

Inspired by [Kubernetes' Prow PR Size plugin](https://prow.k8s.io/plugins), this GitHub Action 
will automatically label a PR based on the number of lines changed, excluding files
marked as `linguist-generated` in `.gitattributes`.

- [Usage](#usage)
- [How it works](#how-it-works)
- [Principles](#principles)
  - [Declarative configuration](#declarative-configuration)
- [Credits](#credits)


## Usage

Create a `.github/workflows/size-labeler.yml` file in your repository with the following contents:

```yaml
name: Size Labeler
on: [pull_request]

permissions:
  contents: read
  issues: write
  pull-requests: write

jobs:
  labeler:
    runs-on: [ubuntu-latest]
    steps:
    - uses: actions/checkout@v3
    - name: Labeler action
      uses: ./.
      with:
        repo-token: ${{ secrets.GITHUB_TOKEN }}
```

Then define your configuration in a `.github/pr-size-labeler.yml` file in your repository with the following contents:

```yaml
labels:
- name: size/XS
  color: '009900'
  min-lines: 0
  description: 'Denotes a PR that changes 0-9 lines'
- name: size/S
  color: '0077bb'
  min-lines: 10
  description: 'Denotes a PR that changes 10-29 lines'
- name: size/M
  color: 'eebb00'
  min-lines: 31
  description: 'Denotes a PR that changes 31-99 lines'
- name: size/L
  color: 'ee9900'
  min-lines: 100
  description: 'Denotes a PR that changes 100-499 lines'
- name: size/XL
  color: 'ee5500'
  min-lines: 500
  description: 'Denotes a PR that changes 500-999 lines'
- name: size/XXL
  color: 'ee0000'
  min-lines: 1000
  description: 'Denotes a PR that changes 1000+ lines'
```

## How it works

When installed as above, the action will run on every PR and will:
* Create/update the labels defined in the `.github/pr-size-labeler.yml` file in GitHub. If you change the minimum number of lines for a label, it will not go back and update open/closed PRs based on the new calculated sizes. If you push a new commit though, it will update the label on the PR.
* Will add/update the size label appropriately on the PR based on the number of lines changed in the PR.
* If you have a `.gitattributes` file in your repository, it will exclude files marked as `linguist-generated` from the line count in determining which label to apply.

## Principles

### Declarative configuration

Size labels are defined declaratively in a YAML file in the repository and will be created/updated in GitHub when the action runs. This allows for easy management of label names, colors, minimum line counts, and descriptions.


## Credits

* This action was inspired by [Kubernetes' Prow PR Size plugin](https://prow.k8s.io/plugins) and uses its gitattributes parsing code.
* The pattern used for packaging a Go GitHub Action with a javascript shim can be found [here](https://full-stack.blend.com/how-we-write-github-actions-in-go.html#small-entrypoint-scripts)
