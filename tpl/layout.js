const api = (() => {

    const req = async (method, path, data) => {
        let body;
        if (data) {
            body = JSON.stringify(data);
        }
        const res = await fetch(path, {
            method,
            body,
            credentials: 'include',
            headers: {
                'content-type': 'application/json',
            },
        });
        const json = await res.json();
        if (res.status !== 200) {
            throw new Error(json.message);
        }
        return json;
    };

    return {
        get: (path) => req('GET', path),
        post: (path, data) => req('POST', path, data),
    };

})();

const gid = (id) => document.getElementById(id);

const setDisabled = (disabled, ...elements) => elements.forEach(e => e.disabled = disabled);

const redirect = () => {
    location.href = `${location.protocol}//${sessionDomain}${location.port ? `:${location.port}` : ''}`;
};