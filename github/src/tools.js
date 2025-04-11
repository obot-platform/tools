import { GPTScript } from "@gptscript-ai/gptscript";

export async function searchIssuesAndPRs(octokit, owner, repo, query, perPage = 100, page = 1) {
    let q = '';

    if (owner) {
        const { data: { type } } = await octokit.users.getByUsername({ username: owner });
        const ownerQualifier = type === 'User' ? `user:${owner}` : `org:${owner}`;
        q = repo ? `repo:${owner}/${repo}` : ownerQualifier;
    } else if (repo) {
        throw new Error('Repository given without an owner. Please provide an owner.');
    } else {
        throw new Error('Owner and repository must be provided.');
    }

    if (query) {
        q += ` ${query}`;
    }

    const { data: { items } } = await octokit.search.issuesAndPullRequests({
        q: q.trim(),
        per_page: perPage,
        page: page
    });

    try {
        const gptscriptClient = new GPTScript();
        const elements = items.map(issue => {
            return {
                name: `${issue.id}`,
                description: '',
                contents: `#${issue.number} - ${issue.title} (ID: ${issue.id}) - ${issue.html_url}`
            }
        });
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${query}_github_issues_prs`,
            description: `Search results for ${query} on GitHub`
        })

        console.log(`Created dataset with ID ${datasetID} with ${elements.length} results`);
    } catch (e) {
        console.log('Failed to create dataset:', e)
    }
}

export async function getIssue(octokit, owner, repo, issueNumber) {
    const { data } = await octokit.issues.get({
        owner,
        repo,
        issue_number: issueNumber,
    });
    console.log(data);
    console.log(`https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function createIssue(octokit, owner, repo, title, body) {
    const issue = await octokit.issues.create({
        owner,
        repo,
        title,
        body
    });

    console.log(`Created issue #${issue.data.number} - ${issue.data.title} (ID: ${issue.data.id}) - https://github.com/${owner}/${repo}/issues/${issue.data.number}`);
}

export async function modifyIssue(octokit, owner, repo, issueNumber, title, body) {
    const issue = await octokit.issues.update({
        owner,
        repo,
        issue_number: issueNumber,
        title,
        body
    });

    console.log(`Modified issue #${issue.data.number} - ${issue.data.title} (ID: ${issue.data.id}) - https://github.com/${owner}/${repo}/issues/${issue.data.number}`);
}

export async function closeIssue(octokit, owner, repo, issueNumber) {
    await octokit.issues.update({
        owner,
        repo,
        issue_number: issueNumber,
        state: 'closed'
    });
    console.log(`Closed issue #${issueNumber} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function listIssueComments(octokit, owner, repo, issueNumber) {
    const { data } = await octokit.issues.listComments({
        owner,
        repo,
        issue_number: issueNumber,
    });

    try {
        const gptscriptClient = new GPTScript();
        const elements = data.map(comment => {
            return {
                name: `${comment.id}`,
                description: '',
                contents: `Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/issues/${issueNumber}#issuecomment-${comment.id}`
            }
        });
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${owner}_${repo}_issue_${issueNumber}_comments`,
            description: `Comments for issue #${issueNumber} in ${owner}/${repo}`
        })
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} comments`);
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}

export async function addCommentToIssue(octokit, owner, repo, issueNumber, comment) {
    const issueComment = await octokit.issues.createComment({
        owner,
        repo,
        issue_number: issueNumber,
        body: comment
    });

    console.log(`Added comment to issue #${issueNumber}: ${issueComment.data.body} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function getPR(octokit, owner, repo, prNumber) {
    const { data } = await octokit.pulls.get({
        owner,
        repo,
        pull_number: prNumber,
    });
    console.log(data);
    console.log(`https://github.com/${owner}/${repo}/pull/${prNumber}`);
}

export async function createPR(octokit, owner, repo, title, body, head, base) {
    const pr = await octokit.pulls.create({
        owner,
        repo,
        title,
        body,
        head,
        base
    });

    console.log(`Created PR #${pr.data.number} - ${pr.data.title} (ID: ${pr.data.id}) - https://github.com/${owner}/${repo}/pull/${pr.data.number}`);
}

export async function modifyPR(octokit, owner, repo, prNumber, title, body) {
    const pr = await octokit.pulls.update({
        owner,
        repo,
        pull_number: prNumber,
        title,
        body
    });

    console.log(`Modified PR #${pr.data.number} - ${pr.data.title} (ID: ${pr.data.id}) - https://github.com/${owner}/${repo}/pull/${pr.data.number}`);
}

export async function closePR(octokit, owner, repo, prNumber) {
    await octokit.pulls.update({
        owner,
        repo,
        pull_number: prNumber,
        state: 'closed'
    });

    console.log(`Deleted PR #${prNumber} - https://github.com/${owner}/${repo}/pull/${prNumber}`);
}

export async function listPRComments(octokit, owner, repo, prNumber) {
    const { data } = await octokit.issues.listComments({
        owner,
        repo,
        issue_number: prNumber,
    });

    try {
        const gptscriptClient = new GPTScript();
        const elements = data.map(comment => {
            return {
                name: `${comment.id}`,
                description: '',
                contents: `Comment by ${comment.user.login}: ${comment.body} - https://github.com/${owner}/${repo}/pull/${prNumber}#issuecomment-${comment.id}`
            }
        });
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${owner}_${repo}_pr_${prNumber}_comments`,
            description: `Comments for PR #${prNumber} in ${owner}/${repo}`
        })
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} comments`);
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}

export async function addCommentToPR(octokit, owner, repo, prNumber, comment) {
    const prComment = await octokit.issues.createComment({
        owner,
        repo,
        issue_number: prNumber,
        body: comment
    });

    console.log(`Added comment to PR #${prNumber}: ${prComment.data.body} - https://github.com/${owner}/${repo}/pull/${prNumber}`);
}

// Private helper functions for repository listing
async function _fetchUserRepos(octokit, username) {
    const allRepos = [];
    for await (const response of octokit.paginate.iterator(octokit.repos.listForUser, {
        username,
        per_page: 100
    })) {
        allRepos.push(...response.data);
    }
    return allRepos;
}

async function _fetchAuthenticatedUserRepos(octokit) {
    const allRepos = [];
    for await (const response of octokit.paginate.iterator(octokit.repos.listForAuthenticatedUser, {
        per_page: 100
    })) {
        allRepos.push(...response.data);
    }
    return allRepos;
}

async function _fetchOrgRepos(octokit, org) {
    const allRepos = [];
    for await (const response of octokit.paginate.iterator(octokit.repos.listForOrg, {
        org,
        per_page: 100
    })) {
        allRepos.push(...response.data);
    }
    return allRepos;
}

export async function listRepos(octokit, owner) {
    try {
        let allRepos = [];
        let description = "";
        let datasetName = "";
        const authenticatedUser = await octokit.users.getAuthenticated();
        const username = authenticatedUser.data.login;

        // Determine what type of repository listing we need to perform
        if (!owner) {
            // List authenticated user's repos
            allRepos = await _fetchAuthenticatedUserRepos(octokit);
            description = `GitHub repos for authenticated user ${username}`;
            datasetName = "my_github_repos";
            console.log(`Found ${allRepos.length} repositories for authenticated user ${username}`);
        } else {
            try {
                // Check if the owner is a user or an organization
                const { data: ownerData } = await octokit.users.getByUsername({ username: owner });

                if (ownerData.type === 'Organization') {
                    // List organization repos
                    allRepos = await _fetchOrgRepos(octokit, owner);
                    description = `GitHub repos for organization ${owner}`;
                    datasetName = `${owner}_github_repos`;
                    console.log(`Found ${allRepos.length} repositories for organization ${owner}`);
                } else {
                    // List user repos
                    allRepos = await _fetchUserRepos(octokit, owner);
                    description = `GitHub repos for user ${owner}`;
                    datasetName = `${owner}_github_repos`;
                    console.log(`Found ${allRepos.length} repositories for user ${owner}`);
                }
            } catch (error) {
                if (error.status === 404) {
                    // If not found as user, try as organization
                    try {
                        allRepos = await _fetchOrgRepos(octokit, owner);
                        description = `GitHub repos for organization ${owner}`;
                        datasetName = `${owner}_github_repos`;
                        console.log(`Found ${allRepos.length} repositories for organization ${owner}`);
                    } catch (orgError) {
                        throw new Error(`Could not find repos for '${owner}' as either a user or organization`);
                    }
                } else {
                    throw error;
                }
            }
        }

        if (allRepos.length > 0) {
            const gptscriptClient = new GPTScript();
            const elements = allRepos.map(repo => {
                const repoOwner = repo.owner ? repo.owner.login : owner || username;
                return {
                    name: `${repo.id}`,
                    description: '',
                    contents: `${repoOwner}/${repo.name} (ID: ${repo.id}) - ${repo.description || 'No description'} - https://github.com/${repoOwner}/${repo.name}`
                };
            });

            const datasetID = await gptscriptClient.addDatasetElements(elements, {
                name: datasetName,
                description
            });
            console.log(`Created dataset with ID ${datasetID} with ${elements.length} repositories`);
        } else {
            console.log('No repositories found');
        }
    } catch (e) {
        console.log('Failed to list repositories:', e);
    }
}

export async function getStarCount(octokit, owner, repo) {
    const { data } = await octokit.repos.get({
        owner,
        repo,
    });
    console.log(data.stargazers_count);
}

export async function listAssignedIssues(octokit) {
    const user = await octokit.rest.users.getAuthenticated();

    const { data } = await octokit.rest.search.issuesAndPullRequests({
        q: `is:open is:issue assignee:${user.data.login} archived:false`
    });

    try {
        const gptscriptClient = new GPTScript();
        const elements = data.items.map(issue => {
            const owner = issue.html_url.split('/')[3]
            const repo = issue.html_url.split('/')[4]
            return {
                name: `${issue.id}`,
                description: '',
                contents: `${owner}/${repo} #${issue.number} - ${issue.title} (ID: ${issue.id}) - ${issue.html_url}`
            }
        });

        if (elements.length > 0) {
            const datasetID = await gptscriptClient.addDatasetElements(elements, {
                name: `assigned_issues`,
                description: `Issues assigned to the authenticated user`
            });
            console.log(`Created dataset with ID ${datasetID} with ${elements.length} issues`);
        } else {
            console.log('No assigned issues found');
        }
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}

export async function listPRsForReview(octokit) {
    const user = await octokit.rest.users.getAuthenticated();

    const { data } = await octokit.rest.search.issuesAndPullRequests({
        q: `is:pr review-requested:${user.data.login} is:open archived:false`,
    });

    try {
        const gptscriptClient = new GPTScript();
        const elements = data.items.map(pr => {
            const owner = pr.html_url.split('/')[3]
            const repo = pr.html_url.split('/')[4]
            return {
                name: `${pr.id}`,
                description: '',
                contents: `${owner}/${repo} #${pr.number} - ${pr.title} (ID: ${pr.id}) - ${pr.html_url}`
            }
        });

        if (elements.length > 0) {
            const datasetID = await gptscriptClient.addDatasetElements(elements, {
                name: `pr_review_requests`,
                description: `PRs requesting review from the authenticated user`
            });
            console.log(`Created dataset with ID ${datasetID} with ${elements.length} PRs`);
        } else {
            console.log('No PRs requesting review found');
        }
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}

export async function addIssueLabels(octokit, owner, repo, issueNumber, labels) {
    const response = await octokit.issues.addLabels({
        owner,
        repo,
        issue_number: issueNumber,
        labels: labels.split(',').map(label => label.trim())
    });

    console.log(`Added labels to issue #${issueNumber}: ${response.data.map(label => label.name).join(', ')} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function removeIssueLabels(octokit, owner, repo, issueNumber, labels) {
    // If labels is empty or undefined, remove all labels
    if (!labels) {
        await octokit.issues.removeAllLabels({
            owner,
            repo,
            issue_number: issueNumber
        });
        console.log(`Removed all labels from issue #${issueNumber} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
        return;
    }

    // Remove specific labels
    const labelArray = labels.split(',').map(label => label.trim());
    for (const label of labelArray) {
        await octokit.issues.removeLabel({
            owner,
            repo,
            issue_number: issueNumber,
            name: label
        });
    }
    console.log(`Removed labels from issue #${issueNumber}: ${labelArray.join(', ')} - https://github.com/${owner}/${repo}/issues/${issueNumber}`);
}

export async function getUser(octokit) {
    await octokit.users.getAuthenticated();
}

export async function getJobLogs(octokit, owner, repo, jobId) {
    const response = await octokit.request('GET /repos/{owner}/{repo}/actions/jobs/{job_id}/logs', {
        owner,
        repo,
        job_id: jobId
    });
    console.log(response.data);
}

export async function listUserOrgs(octokit) {
    try {
        const allOrgs = [];

        // Use Octokit pagination to get all organizations
        for await (const response of octokit.paginate.iterator(octokit.orgs.listForAuthenticatedUser, {
            per_page: 100
        })) {
            allOrgs.push(...response.data);
        }

        console.log(`Found ${allOrgs.length} organizations for authenticated user`);

        const gptscriptClient = new GPTScript();
        const elements = allOrgs.map(org => {
            return {
                name: `${org.id}`,
                description: '',
                contents: `${org.login} (ID: ${org.id}) - https://github.com/${org.login}`
            }
        });

        if (elements.length > 0) {
            const datasetID = await gptscriptClient.addDatasetElements(elements, {
                name: `user_github_orgs`,
                description: `GitHub organizations for authenticated user`
            });
            console.log(`Created dataset with ID ${datasetID} with ${elements.length} organizations`);
        } else {
            console.log('No organizations found for authenticated user');
        }
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}
