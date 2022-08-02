/**
 * Gets all commit messages of a push or title and body of a pull request
 * concatenated to one message.
 *
 * Uses the following environment variables:
 *    ignoreTitle (default: 'true') — Don't use PR title
 *    ignoreDescription (default: 'true') — Don't use the Pull Request description
 *    ignoreLatestCommitMessage (default: 'false') — Use the latest commit message content
 *
 * @param {object} inputs
 * @param {object} inputs.github - A pre-authenticated octokit/rest.js client with pagination plugins.
 * @param {object} inputs.context - An object containing the context of the workflow run.
 * @param {object} inputs.core - A reference to the '@actions/core' package.
 * @returns string[]
 */
getCommitMessages = async ({github, context, core}) => {
  const messages = []
  let [ignoreTitle, ignoreDescription, ignoreLatestCommitMessage] = [true, true, false]
  ignoreTitle = "ignoreTitle" in process.env ? process.env.ignoreTitle === 'true' : ignoreTitle ;
  ignoreDescription = "ignoreDescription" in process.env ? process.env.ignoreDescription === 'true' : ignoreDescription ;
  ignoreLatestCommitMessage = "ignoreLatestCommitMessage" in process.env ? process.env.ignoreLatestCommitMessage === 'true' : ignoreLatestCommitMessage ;

  switch (context.eventName) {
    case 'pull_request': {
      if (!context.payload) {
        throw new Error('No payload found in the context.')
      }

      if (!context.payload.pull_request) {
        throw new Error('No pull_request found in the payload.')
      }

      let message = ''

      if (!ignoreTitle) {
        if (!context.payload.pull_request.title) {
          throw new Error('No title found in the pull_request.')
        }

        message += context.payload.pull_request.title
      } else {
        core.debug(' - skipping title')
      }

      if (!ignoreDescription) {
        if (context.payload.pull_request.body) {
          message = message.concat(
            message !== '' ? '\n\n' : '',
            context.payload.pull_request.body
          )
        }
      } else {
        core.debug(' - skipping description')
      }

      if (message) {
        messages.push(message)
      }

      if (!ignoreLatestCommitMessage) {
        if (!context.payload.pull_request.number) {
          throw new Error('No number found in the pull_request.')
        }

        if (!context.payload.repository) {
          throw new Error('No repository found in the payload.')
        }

        if (!context.payload.repository.name) {
          throw new Error('No name found in the repository.')
        }

        if (
          !context.payload.repository.owner ||
          (!context.payload.repository.owner.login &&
            !context.payload.repository.owner.name)
        ) {
          throw new Error('No owner found in the repository.')
        }

        const commitMessages = await getCommitMessagesFromPullRequest({github, context, core})
        const message = commitMessages[commitMessages.length - 1]
        messages.push(messages.length > 0 ? ''.concat('\n\n', message) : message)
      } else {
        core.debug(' - skipping commit message')
      }

      break
    }
    case 'push': {
      if (!context.payload) {
        throw new Error('No payload found in the context.')
      }

      if (
        !context.payload.commits ||
        !context.payload.commits.length
      ) {
        core.debug(' - skipping commits')
        break
      }

      const message =
        context.payload.commits[
          context.payload.commits.length - 1
        ].message
      messages.push(message)

      break
    }
    default: {
      throw new Error(`Event "${context.eventName}" is not supported.`)
    }
  }

  return messages
}

async function getCommitMessagesFromPullRequest({github, context, core}) {
  const repositoryOwner = context.payload.repository.owner.login
  const repositoryName = context.payload.repository.name
  const pullRequestNumber = context.payload.pull_request.number

  core.debug('Get messages from pull request...')
  core.debug(` - repositoryOwner: ${repositoryOwner}`)
  core.debug(` - repositoryName: ${repositoryName}`)
  core.debug(` - pullRequestNumber: ${pullRequestNumber}`)

  const query = `
  query commitMessages(
    $repositoryOwner: String!
    $repositoryName: String!
    $pullRequestNumber: Int!
    $numberOfCommits: Int = 100
  ) {
    repository(owner: $repositoryOwner, name: $repositoryName) {
      pullRequest(number: $pullRequestNumber) {
        commits(last: $numberOfCommits) {
          edges {
            node {
              commit {
                message
              }
            }
          }
        }
      }
    }
  }
`
  const variables = {
    repositoryOwner: repositoryOwner,
    repositoryName: repositoryName,
    pullRequestNumber: pullRequestNumber,
  }

  core.debug(` - query: ${query}`)
  core.debug(` - variables: ${JSON.stringify(variables, null, 2)}`)

  const {repository} = await github.graphql(query, variables)

  core.debug(` - response: ${JSON.stringify(repository, null, 2)}`)

  let messages = []

  if (repository.pullRequest) {
    messages = repository.pullRequest.commits.edges.map( edgeItem => {
      return edgeItem.node.commit.message
    })
  }

  return messages
}

module.exports.run = async ({ github, context, core}) => {
  try {
    const messages = await getCommitMessages({ github, context, core })
    let commitMessage = ''
    if (messages && messages.length === 0) {
      core.info('No commits found in the payload, skipping check.')
    } else {
      commitMessage = messages.join('\n').replace(/\"/gi, '\\"')
      core.info(`Commit messages found:\n ${messages}`)
    }
    core.setOutput('message', commitMessage)
  } catch (error) {
    core.setFailed(error)
  }
}
