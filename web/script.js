// GQL to Neo4j Converter Web UI with WASM backend

// Load example GQL schema from file
async function loadExampleSchema() {
    try {
        const response = await fetch('example-schema.gql');
        if (!response.ok) {
            throw new Error(`Failed to load example schema: ${response.status}`);
        }
        return await response.text();
    } catch (error) {
        console.error('Error loading example schema:', error);
        // Fallback to a minimal example if the file can't be loaded
        return `CREATE PROPERTY GRAPH socialNetworkType {
    (:Person {
        givenName :: STRING(2, 50) NOT NULL,
        familyName :: VARCHAR(50) NOT NULL,
        age :: INT8 NOT NULL
    }) PRIMARY KEY (givenName, familyName)
}`;
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    const gqlInput = document.getElementById('gql-input');
    const neo4jOutput = document.getElementById('neo4j-output');
    const errorMessage = document.getElementById('error-message');
    const neo4jVersion = document.getElementById('neo4j-version');
    const neo4jEdition = document.getElementById('neo4j-edition');
    const apocEnabled = document.getElementById('apoc-enabled');
    const loadingIndicator = document.getElementById('loading-indicator');
    const loadingText = document.querySelector('.loading-text');

    // Button elements
    const loadExampleBtn = document.getElementById('load-example');
    const clearInputBtn = document.getElementById('clear-input');
    const copyOutputBtn = document.getElementById('copy-output');
    const downloadOutputBtn = document.getElementById('download-output');

    let debounceTimeout;
    let converter = null;

    // Initialize WASM converter
    try {
        showLoading('Initializing converter...');
        converter = new WASMConverter();
        await converter.load();
        hideLoading();
        showToast('Converter ready!', 'success');
        console.log('WASM converter ready');

        // Update version info in footer
        const versionElement = document.getElementById('version-info');
        if (versionElement) {
            versionElement.textContent = converter.getVersion();
        }
    } catch (error) {
        console.error('Failed to initialize WASM converter:', error);
        await updateOutput('', `Failed to initialize converter: ${error.message}`);
        hideLoading();

        // Update version info to show error state
        const versionElement = document.getElementById('version-info');
        if (versionElement) {
            versionElement.textContent = 'Version unavailable';
        }
        return;
    }

    // Add event listeners
    [gqlInput, neo4jVersion, apocEnabled].forEach(input => {
        input.addEventListener('input', () => {
            clearTimeout(debounceTimeout);
            debounceTimeout = setTimeout(convertGQL, 500);
        });
    });

    // Add change listener for select dropdown
    neo4jEdition.addEventListener('change', () => {
        clearTimeout(debounceTimeout);
        debounceTimeout = setTimeout(convertGQL, 500);
    });

    // Button event listeners
    loadExampleBtn.addEventListener('click', async () => {
        try {
            showLoading('Loading example...');
            const exampleSchema = await loadExampleSchema();
            gqlInput.value = exampleSchema;
            hideLoading();
            convertGQL();
            showToast('Example loaded', 'success');
        } catch (error) {
            hideLoading();
            showToast('Failed to load example', 'error');
            console.error('Error loading example:', error);
        }
    });

    clearInputBtn.addEventListener('click', () => {
        gqlInput.value = '';
        updateOutput('Neo4j conversion will appear here...');
        showToast('Input cleared', 'success');
    });

    copyOutputBtn.addEventListener('click', async () => {
        const outputText = neo4jOutput.textContent;
        if (!outputText || outputText === 'Neo4j conversion will appear here...') {
            showToast('No output to copy', 'error');
            return;
        }

        try {
            await navigator.clipboard.writeText(outputText);
            showToast('Output copied to clipboard!', 'success');
        } catch (error) {
            console.error('Failed to copy:', error);
            showToast('Failed to copy to clipboard', 'error');
        }
    });

    downloadOutputBtn.addEventListener('click', () => {
        const outputText = neo4jOutput.textContent;
        if (!outputText || outputText === 'Neo4j conversion will appear here...') {
            showToast('No output to download', 'error');
            return;
        }

        const blob = new Blob([outputText], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'neo4j-constraints.cypher';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        showToast('File downloaded!', 'success');
    });

    function showLoading(message = 'Converting...') {
        if (loadingIndicator) {
            if (loadingText) {
                loadingText.textContent = message;
            }
            loadingIndicator.style.display = 'flex';
        }
    }

    function hideLoading() {
        if (loadingIndicator) {
            loadingIndicator.style.display = 'none';
        }
    }

    function showToast(message, type = 'success') {
        const toast = document.getElementById('toast');
        const toastMessage = document.getElementById('toast-message');

        if (toast && toastMessage) {
            toastMessage.textContent = message;
            toast.className = `toast ${type}`;
            toast.classList.add('show');

            setTimeout(() => {
                toast.classList.remove('show');
            }, 3000);
        }
    }

    /**
     * @param {string} text
     * @param {string | boolean} error
     * @returns {Promise<void>}
     */
    async function updateOutput(text, error = false) {
        if (error) {
            errorMessage.textContent = error;
            errorMessage.classList.remove('hidden');
            neo4jOutput.textContent = '';
        } else {
            errorMessage.classList.add('hidden');
            neo4jOutput.textContent = text || 'No conversion result';
            // Re-highlight the syntax
            if (window.Prism && text) {
                Prism.highlightElement(neo4jOutput);
            }
        }
    }

    /**
     * @returns {Promise<void>}
     */
    async function convertGQL() {
        const gql = gqlInput.value.trim();

        if (!gql) {
            await updateOutput('Neo4j conversion will appear here...');
            return;
        }

        if (!converter) {
            await updateOutput('', 'Converter not initialized');
            return;
        }

        try {
            showLoading('Converting...');
            const result = await converter.convert(
                gql,
                neo4jVersion.value || 'latest',
                neo4jEdition.value || 'enterprise',
                apocEnabled.checked
            );

            hideLoading();

            if (result.error) {
                await updateOutput('', result.error);
            } else {
                await updateOutput(result.neo4j);
            }
        } catch (error) {
            hideLoading();
            console.error('Conversion error:', error);
            await updateOutput('', `Failed to convert: ${error.message}`);
        }
    }

    // Add keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        // Ctrl/Cmd + Enter to convert
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            e.preventDefault();
            convertGQL();
        }

        // Ctrl/Cmd + E to load example
        if ((e.ctrlKey || e.metaKey) && e.key === 'e') {
            e.preventDefault();
            loadExampleBtn.click();
        }
    });
});
