(() => {
    const
        gid = id => document.getElementById(id),
        f = {
            gender: gid('gender'),
            last_name: gid('last_name'),
            first_name: gid('first_name'),
            middle_name: gid('middle_name'),
            email: gid('email'),
            password: gid('password'),
            confirm_password: gid('confirm_password'),
            submit_button: gid('submit_button'),
        };

    const filter = (v) => ![
        'submit_button',
        'confirm_password'
    ].includes(v);

    const getValues = () => Object
        .keys(f)
        .filter(filter)
        .reduce((obj, v) => {
            obj[v] = f[v].value;
            return obj;
        }, Object.create(null));

    const setDisabled = (disabled) => {
        Object
            .keys(f)
            .forEach(v => {
                f[v].disabled = disabled;
            });
    };

    const validate = () => {
        if (f.password.value !== f.confirm_password.value) {
            f.confirm_password.setCustomValidity('Пароли не совпадают');
            return false;
        }
        f.confirm_password.setCustomValidity('');
        return true;
    };

    gid('form').addEventListener('submit', async (event) => {
        event.preventDefault();
        if (!validate()) {
            return;
        }
        setDisabled(true);
        try {
            const res = await fetch('/api/register', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'content-type': 'application/json',
                },
                body: JSON.stringify(getValues()),
            });
            const json = await res.json();
            if (res.status !== 200) {

                return;
            }
            redirect();
        } catch (err) {
            console.error(err);
        }
        // setDisabled(false);
        console.log(getValues());
    });

    const passwordValidator = () => {
        validate()
    };

    f.password.addEventListener('change', passwordValidator);
    f.confirm_password.addEventListener('keyup', passwordValidator);
})();