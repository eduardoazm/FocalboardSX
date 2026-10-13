// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
import React, {useState} from 'react'
import {Link} from 'react-router-dom'
import {FormattedMessage, useIntl} from 'react-intl'

import Button from '../widgets/buttons/button'
import client from '../octoClient'
import './changePasswordPage.scss'
import {IUser} from '../user'
import {useAppSelector} from '../store/hooks'
import {getMe} from '../store/users'

const ChangePasswordPage = () => {
    const intl = useIntl()
    const [oldPassword, setOldPassword] = useState('')
    const [newPassword, setNewPassword] = useState('')
    const [errorMessage, setErrorMessage] = useState('')
    const [succeeded, setSucceeded] = useState(false)
    const user = useAppSelector<IUser|null>(getMe)

    if (!user) {
        return (
            <div className='ChangePasswordPage'>
                <div className='title'>
                    <FormattedMessage id='changePassword.title' defaultMessage='Alterar Senha'/>
                </div>
                <Link to='/login'>
                    <FormattedMessage id='changePassword.login-first' defaultMessage='Faça login primeiro'/>
                </Link>
            </div>
        )
    }

    const handleSubmit = async (userId: string): Promise<void> => {
        const response = await client.changePassword(userId, oldPassword, newPassword)
        if (response.code === 200) {
            setOldPassword('')
            setNewPassword('')
            setErrorMessage('')
            setSucceeded(true)
        } else {
            setErrorMessage(`Change password failed: ${response.json?.error}`)
        }
    }

    return (
        <div className='ChangePasswordPage'>
            <div className='title'>
                <FormattedMessage id='changePassword.title' defaultMessage='Alterar Senha'/>
            </div>
            <form
                onSubmit={(e: React.FormEvent) => {
                    e.preventDefault()
                    handleSubmit(user.id)
                }}
            >
                <div className='oldPassword'>
                    <input
                        id='login-oldpassword'
                        type='password'
                        placeholder={intl.formatMessage({id: 'changePassword.current-password-placeholder', defaultMessage: 'Digite a senha atual'})}
                        value={oldPassword}
                        onChange={(e) => {
                            setOldPassword(e.target.value)
                            setErrorMessage('')
                        }}
                    />
                </div>
                <div className='newPassword'>
                    <input
                        id='login-newpassword'
                        type='password'
                        placeholder={intl.formatMessage({id: 'changePassword.new-password-placeholder', defaultMessage: 'Digite a nova senha'})}
                        value={newPassword}
                        onChange={(e) => {
                            setNewPassword(e.target.value)
                            setErrorMessage('')
                        }}
                    />
                </div>
                <Button
                    filled={true}
                    submit={true}
                >
                    <FormattedMessage id='changePassword.submit-button' defaultMessage='Alterar senha'/>
                </Button>
            </form>
            {errorMessage &&
                <div className='error'>
                    {errorMessage}
                </div>
            }
            {succeeded &&
                <Link
                    className='succeeded'
                    to='/'
                >
                    <FormattedMessage id='changePassword.success' defaultMessage='Senha alterada com sucesso, clique para continuar.'/>
                </Link>
            }
            {!succeeded &&
                <Link to='/'>
                    <FormattedMessage id='changePassword.cancel' defaultMessage='Cancelar'/>
                </Link>
            }
        </div>
    )
}

export default React.memo(ChangePasswordPage)
