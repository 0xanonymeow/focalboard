// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
import React, {useState, useEffect} from 'react'
import {useHistory, Link, Redirect} from 'react-router-dom'
import {FormattedMessage} from 'react-intl'

import {useAppDispatch, useAppSelector} from '../store/hooks'
import {fetchMe, getLoggedIn} from '../store/users'

import Button from '../widgets/buttons/button'
import client from '../octoClient'
import {Utils} from '../utils'
import './registerPage.scss'

const RegisterPage = () => {
    const [username, setUsername] = useState('')
    const [password, setPassword] = useState('')
    const [email, setEmail] = useState('')
    const [errorMessage, setErrorMessage] = useState('')
    const [isFromInvitation, setIsFromInvitation] = useState(false)
    const history = useHistory()
    const dispatch = useAppDispatch()
    const loggedIn = useAppSelector<boolean|null>(getLoggedIn)

    // Check for invitation email and pre-fill
    useEffect(() => {
        const invitationEmail = localStorage.getItem('invitation_email')
        if (invitationEmail) {
            setEmail(invitationEmail)
            setIsFromInvitation(true)
        }
    }, [])

    const handleRegister = async (): Promise<void> => {
        // Validate email matches invitation if coming from invitation
        const invitationEmail = localStorage.getItem('invitation_email')
        if (invitationEmail && email !== invitationEmail) {
            setErrorMessage('Email must match the invitation email address')
            return
        }

        const queryString = new URLSearchParams(window.location.search)
        const signupToken = queryString.get('t') || ''

        const response = await client.register(email, username, password, signupToken)
        if (response.code === 200) {
            const logged = await client.login(username, password)
            if (logged) {
                await dispatch(fetchMe())
                
                // Check for stored invitation token
                const invitationToken = localStorage.getItem('invitation_token')
                if (invitationToken) {
                    localStorage.removeItem('invitation_token')
                    localStorage.removeItem('invitation_email')
                    history.push(`/invite/${invitationToken}`)
                    return
                }
                
                history.push('/')
            }
        } else if (response.code === 401) {
            setErrorMessage('Invalid registration link, please contact your administrator')
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
                        placeholder={isFromInvitation ? 'Email (from invitation)' : 'Enter email'}
                        value={email}
                        onChange={(e) => setEmail(e.target.value.trim())}
                        readOnly={isFromInvitation}
                        style={{
                            backgroundColor: isFromInvitation ? 'rgb(var(--center-channel-bg-rgb))' : undefined,
                            cursor: isFromInvitation ? 'not-allowed' : undefined
                        }}
                    />
                    {isFromInvitation && (
                        <div className='email-locked-message'>
                            <FormattedMessage
                                id='register.email-locked'
                                defaultMessage='Email is pre-filled from your invitation'
                            />
                        </div>
                    )}
                </div>
                <div className='username'>
                    <input
                        id='login-username'
                        placeholder={'Enter username'}
                        value={username}
                        onChange={(e) => setUsername(e.target.value.trim())}
                    />
                </div>
                <div className='password'>
                    <input
                        id='login-password'
                        type='password'
                        placeholder={'Enter password'}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                    />
                </div>
                <Button
                    filled={true}
                    submit={true}
                >
                    {'Register'}
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
