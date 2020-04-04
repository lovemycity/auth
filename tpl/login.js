(() => {
    const
        f = {
            email: gid('email'),
            password: gid('password'),
        },
        button = gid('button'),
        error = gid('error');

    const _getValues = () => Object
        .keys(f)
        .reduce((o, v) => {
            o[v] = f[v].value;
            return o;
        }, Object.create(null));

    const _setError = (s) => error.textContent = s;

    const _setDisabled = (d) => {
        setDisabled(d,
            ...Object.values(f),
            button);
    };

    gid('form').addEventListener('submit', async (event) => {
        event.preventDefault();
        _setError('');
        _setDisabled(true);
        try {
            await api.post('/api/login', _getValues());
            // redirect();
        } catch (err) {
            _setError(err.message);
            _setDisabled(false);
        }
    });
})();