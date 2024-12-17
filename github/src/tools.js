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
    const { repository } = await octokit.graphql(`
        query($owner: String!, $repo: String!, $number: Int!) {
            repository(owner: $owner, name: $repo) {
                issue(number: $number) {
                    id
                    number
                    title
                    url
                    state
                    createdAt
                    updatedAt
                    assignees(first: 10) {
                        nodes {
                            login
                        }
                    }
                    labels(first: 10) {
                        nodes {
                            name
                        }
                    }
                }
            }
        }
    `, {
        owner,
        repo,
        number: parseInt(issueNumber, 10)
    });

    const issue = repository.issue;
    const assignees = issue.assignees.nodes.map(a => a.login).join(', ');
    const labels = issue.labels.nodes.map(l => l.name).join(', ');

    console.log(`#${issue.number} - ${issue.title}
ID: ${issue.id}
State: ${issue.state}
Created: ${issue.createdAt}
Updated: ${issue.updatedAt}
Assignees: ${assignees}
Labels: ${labels}
URL: ${issue.url}`);
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


export async function listRepos(octokit, owner) {
    const repos = await octokit.repos.listForUser({
        username: owner,
        per_page: 100
    });

    try {
        const gptscriptClient = new GPTScript();
        const elements = repos.data.map(repo => {
            return {
                name: `${repo.id}`,
                description: '',
                contents: `${repo.name} (ID: ${repo.id}) - https://github.com/${owner}/${repo.name}`
            }
        });
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${owner}_github_repos`,
            description: `GitHub repos for ${owner}`
        });
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} repositories`);
    } catch (e) {
        console.log('Failed to create dataset:', e);
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

export async function listProjects(octokit, owner) {
    const { data: { type } } = await octokit.users.getByUsername({ username: owner });
    
    let projects;
    if (type === 'User') {
        const { user } = await octokit.graphql(`
            query($username: String!) {
                user(login: $username) {
                    projectsV2(first: 100) {
                        nodes {
                            id
                            title
                            number
                            url
                        }
                    }
                }
            }
        `, {
            username: owner
        });
        projects = user.projectsV2.nodes;
    } else {
        const { organization } = await octokit.graphql(`
            query($org: String!) {
                organization(login: $org) {
                    projectsV2(first: 100) {
                        nodes {
                            id
                            title
                            number
                            url
                        }
                    }
                }
            }
        `, {
            org: owner
        });
        projects = organization.projectsV2.nodes;
    }

    try {
        const gptscriptClient = new GPTScript();
        const elements = projects.map(project => {
            return {
                name: project.id,
                description: '',
                contents: `${project.title} (#${project.number}) - ${project.url}`
            }
        });
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `${owner}_github_projects`,
            description: `GitHub Projects (V2) for ${type.toLowerCase()} ${owner}`
        });
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} projects`);
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}

export async function getProject(octokit, projectId) {
    const { node } = await octokit.graphql(`
        query($projectId: ID!) {
            node(id: $projectId) {
                ... on ProjectV2 {
                    id
                    title
                    number
                    url
                    fields(first: 100) {
                        nodes {
                            ... on ProjectV2Field {
                                id
                                name
                            }
                            ... on ProjectV2IterationField {
                                id
                                name
                            }
                            ... on ProjectV2SingleSelectField {
                                id
                                name
                                options {
                                    id
                                    name
                                }
                            }
                        }
                    }
                }
            }
        }
    `, {
        projectId
    });

    try {
        const gptscriptClient = new GPTScript();
        const elements = node.fields.nodes.map(field => {
            if (field.options) {
                return {
                    name: field.id,
                    description: 'Single Select Field',
                    contents: `${field.name} - Options: ${field.options.map(opt => opt.name).join(', ')}`
                };
            } else {
                return {
                    name: field.id,
                    description: field.__typename || 'Field',
                    contents: `${field.name}`
                };
            }
        });
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `project_${projectId}_fields`,
            description: `Fields for project ID ${projectId} (${node.title} #${node.number})`
        });
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} fields`);
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}

export async function listCards(octokit, projectId, perPage = 100) {
    const parsedPerPage = parseInt(perPage, 10) || 100;
    let hasNextPage = true;
    let cursor = null;
    let allItems = [];
    let projectTitle, projectNumber;

    while (hasNextPage) {
        const { node } = await octokit.graphql(`
            query($projectId: ID!, $first: Int!, $after: String) {
                node(id: $projectId) {
                    ... on ProjectV2 {
                        id
                        title
                        number
                        url
                        items(first: $first, after: $after) {
                            nodes {
                                id
                                fieldValues(first: 100) {
                                    nodes {
                                        ... on ProjectV2ItemFieldTextValue {
                                            field {
                                                ... on ProjectV2FieldCommon {
                                                    name
                                                }
                                            }
                                            text
                                        }
                                        ... on ProjectV2ItemFieldDateValue {
                                            field {
                                                ... on ProjectV2FieldCommon {
                                                    name
                                                }
                                            }
                                            date
                                        }
                                        ... on ProjectV2ItemFieldSingleSelectValue {
                                            field {
                                                ... on ProjectV2FieldCommon {
                                                    name
                                                }
                                            }
                                            name
                                        }
                                        ... on ProjectV2ItemFieldIterationValue {
                                            field {
                                                ... on ProjectV2FieldCommon {
                                                    name
                                                }
                                            }
                                            title
                                        }
                                    }
                                }
                                content {
                                    ... on Issue {
                                        id
                                        title
                                        number
                                        url
                                        state
                                        createdAt
                                        updatedAt
                                        assignees(first: 10) {
                                            nodes {
                                                login
                                            }
                                        }
                                        labels(first: 10) {
                                            nodes {
                                                name
                                            }
                                        }
                                    }
                                    ... on PullRequest {
                                        id
                                        title
                                        number
                                        url
                                        state
                                        createdAt
                                        updatedAt
                                        assignees(first: 10) {
                                            nodes {
                                                login
                                            }
                                        }
                                        labels(first: 10) {
                                            nodes {
                                                name
                                            }
                                        }
                                    }
                                }
                            }
                            pageInfo {
                                hasNextPage
                                endCursor
                            }
                        }
                    }
                }
            }
        `, {
            projectId,
            first: parsedPerPage,
            after: cursor
        });

        if (!projectTitle) {
            projectTitle = node.title;
            projectNumber = node.number;
        }

        allItems = allItems.concat(node.items.nodes);
        hasNextPage = node.items.pageInfo.hasNextPage;
        cursor = node.items.pageInfo.endCursor;
    }

    try {
        const gptscriptClient = new GPTScript();
        const elements = allItems.map(item => {
            const fieldValues = {};
            item.fieldValues.nodes.forEach(fieldValue => {
                // Skip if field or field.name is undefined
                if (!fieldValue?.field?.name) return;
                
                const fieldName = fieldValue.field.name;
                // Handle different field types
                if ('name' in fieldValue) {
                    fieldValues[fieldName] = fieldValue.name; // Status/Single select fields
                } else if ('text' in fieldValue) {
                    fieldValues[fieldName] = fieldValue.text; // Text fields
                } else if ('date' in fieldValue) {
                    fieldValues[fieldName] = fieldValue.date; // Date fields
                } else if ('title' in fieldValue) {
                    fieldValues[fieldName] = fieldValue.title; // Iteration fields
                }
            });

            const content = item.content;
            const assignees = content ? content.assignees.nodes.map(a => a.login).join(', ') : '';
            const labels = content ? content.labels.nodes.map(l => l.name).join(', ') : '';

            return {
                name: item.id,
                description: '',
                contents: content ? 
                    `${content.title} (#${content.number})
ID: ${item.id}
State: ${content.state}
Created: ${content.createdAt}
Updated: ${content.updatedAt}
Assignees: ${assignees} 
Labels: ${labels}
Fields: ${Object.entries(fieldValues).map(([k,v]) => `${k}: ${v}`).join(', ')}
URL: ${content.url}` :
                    `Draft Item (ID: ${item.id})`
            }
        });
        const datasetID = await gptscriptClient.addDatasetElements(elements, {
            name: `project_${projectId}_items`,
            description: `All items in project ${projectTitle} (#${projectNumber})`
        });
        console.log(`Created dataset with ID ${datasetID} with ${elements.length} items`);
    } catch (e) {
        console.log('Failed to create dataset:', e);
    }
}

export async function createCard(octokit, projectId, contentId, status) {
    const { addProjectV2ItemById } = await octokit.graphql(`
        mutation($projectId: ID!, $contentId: ID!) {
            addProjectV2ItemById(input: {
                projectId: $projectId,
                contentId: $contentId
            }) {
                item {
                    id
                }
            }
        }
    `, {
        projectId,
        contentId
    });

    console.log(`Added item to project. Item ID: ${addProjectV2ItemById.item.id}`);

    if (status) {
        // First get the status field ID and option ID
        const { node: project } = await octokit.graphql(`
            query($projectId: ID!) {
                node(id: $projectId) {
                    ... on ProjectV2 {
                        fields(first: 100) {
                            nodes {
                                ... on ProjectV2SingleSelectField {
                                    id
                                    name
                                    options {
                                        id
                                        name
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `, {
            projectId
        });

        // Find the status field and matching option
        const statusField = project.fields.nodes.find(field => field?.name === 'Status');
        if (!statusField) {
            throw new Error('Status field not found in project');
        }

        const statusOption = statusField.options.find(option => option.name === status);
        if (!statusOption) {
            throw new Error(`Status "${status}" not found. Available options: ${statusField.options.map(o => o.name).join(', ')}`);
        }

        // Update the status
        await octokit.graphql(`
            mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
                updateProjectV2ItemFieldValue(input: {
                    projectId: $projectId,
                    itemId: $itemId,
                    fieldId: $fieldId,
                    value: { singleSelectOptionId: $optionId }
                }) {
                    projectV2Item {
                        id
                    }
                }
            }
        `, {
            projectId,
            itemId: addProjectV2ItemById.item.id,
            fieldId: statusField.id,
            optionId: statusOption.id
        });

        console.log(`Set item status to "${status}"`);
    }
}

export async function moveCard(octokit, projectId, itemId, fieldId, optionId) {
    // First get the field and its options to validate
    const { node: project } = await octokit.graphql(`
        query($projectId: ID!) {
            node(id: $projectId) {
                ... on ProjectV2 {
                    fields(first: 100) {
                        nodes {
                            ... on ProjectV2SingleSelectField {
                                id
                                name
                                options {
                                    id
                                    name
                                }
                            }
                        }
                    }
                }
            }
        }
    `, {
        projectId
    });

    // Find the matching option
    const field = project.fields.nodes.find(f => f?.id === fieldId);
    if (!field || !field.options) {
        throw new Error('Field not found or is not a single select field');
    }

    const option = field.options.find(opt => opt.name === optionId);
    if (!option) {
        throw new Error(`Option "${optionId}" not found. Available options: ${field.options.map(o => o.name).join(', ')}`);
    }

    const { updateProjectV2ItemFieldValue } = await octokit.graphql(`
        mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
            updateProjectV2ItemFieldValue(input: {
                projectId: $projectId,
                itemId: $itemId,
                fieldId: $fieldId,
                value: { singleSelectOptionId: $optionId }
            }) {
                projectV2Item {
                    id
                }
            }
        }
    `, {
        projectId,
        itemId,
        fieldId,
        optionId: option.id
    });

    console.log(`Updated item status. Item ID: ${updateProjectV2ItemFieldValue.projectV2Item.id}`);
}

export async function updateCard(octokit, projectId, itemId, fieldId, value) {
    const { updateProjectV2ItemFieldValue } = await octokit.graphql(`
        mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: String!) {
            updateProjectV2ItemFieldValue(input: {
                projectId: $projectId,
                itemId: $itemId,
                fieldId: $fieldId,
                value: { text: $value }
            }) {
                projectV2Item {
                    id
                }
            }
        }
    `, {
        projectId,
        itemId,
        fieldId,
        value
    });

    console.log(`Updated item field. Item ID: ${updateProjectV2ItemFieldValue.projectV2Item.id}`);
}

export async function deleteCard(octokit, projectId, itemId) {
    await octokit.graphql(`
        mutation($projectId: ID!, $itemId: ID!) {
            deleteProjectV2Item(input: {
                projectId: $projectId,
                itemId: $itemId
            }) {
                deletedItemId
            }
        }
    `, {
        projectId,
        itemId
    });

    console.log(`Deleted item ID: ${itemId}`);
}
