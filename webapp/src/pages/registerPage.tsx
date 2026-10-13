// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
import React, {useState} from 'react'
import {useHistory, Link, Redirect} from 'react-router-dom'
import {FormattedMessage, useIntl} from 'react-intl'

import {useAppDispatch, useAppSelector} from '../store/hooks'
import {fetchMe, getLoggedIn} from '../store/users'

import Button from '../widgets/buttons/button'
import client from '../octoClient'
import './registerPage.scss'

const RegisterPage = () => {
    const intl = useIntl()
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [email, setEmail] = useState('')
    const [errorMessage, setErrorMessage] = useState('')
    const history = useHistory()
    const dispatch = useAppDispatch()
    const loggedIn = useAppSelector<boolean|null>(getLoggedIn)

    const handleRegister = async (): Promise<void> => {
        const queryString = new URLSearchParams(window.location.search)
        const signupToken = queryString.get('t') || ''

        const response = await client.register(email, username, password, signupToken)
        if (response.code === 200) {
            const logged = await client.login(username, password)
            if (logged) {
                await dispatch(fetchMe())
                history.push('/')
            }
        } else if (response.code === 401) {
            setErrorMessage(intl.formatMessage({id: 'register.invalid-link', defaultMessage: 'Link de cadastro inválido, entre em contato com o administrador'}))
        } else {
            setErrorMessage(`${response.json?.error}`)
        }
    }

    if (loggedIn) {
        return <Redirect to={'/'}/>
    }

    return (
        <div className='RegisterPage'>
            <form
                onSubmit={(e: React.FormEvent) => {
                    e.preventDefault()
                    handleRegister()
                }}
            >
                <div className='title'>
                    <FormattedMessage
                        id='register.signup-title'
                        defaultMessage='Sign up for your account'
                    />
                </div>
                <div className='email'>
                    <input
                        id='login-email'
                        placeholder={intl.formatMessage({id: 'register.email-placeholder', defaultMessage: 'Digite seu e-mail'})}
                        value={email}
                        onChange={(e) => setEmail(e.target.value.trim())}
                    />
                </div>
                <div className='username'>
                    <input
                        id='login-username'
                        placeholder={intl.formatMessage({id: 'register.username-placeholder', defaultMessage: 'Digite seu usuário'})}
                        value={username}
                        onChange={(e) => setUsername(e.target.value.trim())}
                    />
                </div>
                <div className='password'>
                    <input
                        id='login-password'
                        type='password'
                        placeholder={intl.formatMessage({id: 'register.password-placeholder', defaultMessage: 'Digite sua senha'})}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                    />
                </div>
                <Button
                    filled={true}
                    submit={true}
                >
                    <FormattedMessage
                        id='register.signup-button'
                        defaultMessage='Cadastrar'
                    />
                </Button>
            </form>
            <Link to='/login'>
                <FormattedMessage
                    id='register.login-button'
                    defaultMessage={'or log in if you already have an account'}
                />
            </Link>
            {errorMessage &&
                <div className='error'>
                    {errorMessage}
                </div>
            }
        </div>
    )
}

export default React.memo(RegisterPage)
