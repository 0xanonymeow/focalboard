// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useEffect} from 'react'
import {useParams, useHistory} from 'react-router-dom'
import {FormattedMessage, useIntl} from 'react-intl'

import {Utils} from '../utils'
import client from '../octoClient'
import {sendFlashMessage} from '../components/flashMessages'
import Button from '../widgets/buttons/button'
import './invitationPage.scss'

interface InvitationInfo {
    boardTitle: string
    email: string
    role: string
    boardId: string
    valid: boolean
}

const InvitationPage = (): JSX.Element => {
    const intl = useIntl()
    const history = useHistory()
    const {token} = useParams<{token: string}>()
    const [invitationInfo, setInvitationInfo] = useState<InvitationInfo | null>(null)
    const [loading, setLoading] = useState(true)
    const [accepting, setAccepting] = useState(false)
    const [error, setError] = useState<string>('')

    useEffect(() => {
        if (!token) {
            setError('Invalid invitation link')
            setLoading(false)
            return
        }

        const loadInvitation = async () => {
            try {
                const info = await client.getInvitation(token)
                setInvitationInfo(info)
            } catch (err: any) {
                console.error('Failed to load invitation:', err)
                if (err.message?.includes('expired')) {
                    setError('This invitation has expired')
                } else if (err.message?.includes('used')) {
                    setError('This invitation has already been used')
                } else {
                    setError('Invalid or expired invitation link')
                }
            } finally {
                setLoading(false)
            }
        }

        loadInvitation()
    }, [token])

    const handleAcceptInvitation = async () => {
        if (!token || !invitationInfo) {
            return
        }

        setAccepting(true)
        try {
            await client.acceptInvitation(token)
            
            sendFlashMessage({
                content: intl.formatMessage({
                    id: 'InvitationPage.accepted',
                    defaultMessage: 'Invitation accepted! Welcome to the board.'
                }),
                severity: 'normal'
            })

            // Redirect to the board
            history.push(`/board/${invitationInfo.boardId}`)
        } catch (err: any) {
            console.error('Failed to accept invitation:', err)
            
            let errorMessage = 'Failed to accept invitation'
            if (err.message?.includes('email does not match')) {
                errorMessage = 'This invitation is for a different email address'
            } else if (err.message?.includes('expired')) {
                errorMessage = 'This invitation has expired'
            } else if (err.message?.includes('used')) {
                errorMessage = 'This invitation has already been used'
            }

            sendFlashMessage({
                content: errorMessage,
                severity: 'high'
            })
        } finally {
            setAccepting(false)
        }
    }

    const handleLogin = () => {
        // Store the invitation token to continue after login
        Utils.setLocalStorage('invitation_token', token || '')
        history.push('/login')
    }

    const handleRegister = () => {
        // Store the invitation token to continue after registration
        Utils.setLocalStorage('invitation_token', token || '')
        history.push('/register')
    }

    if (loading) {
        return (
            <div className='InvitationPage'>
                <div className='invitation-container'>
                    <div className='invitation-loading'>
                        <FormattedMessage
                            id='InvitationPage.loading'
                            defaultMessage='Loading invitation...'
                        />
                    </div>
                </div>
            </div>
        )
    }

    if (error || !invitationInfo) {
        return (
            <div className='InvitationPage'>
                <div className='invitation-container'>
                    <div className='invitation-error'>
                        <h2>
                            <FormattedMessage
                                id='InvitationPage.error-title'
                                defaultMessage='Invalid Invitation'
                            />
                        </h2>
                        <p>{error}</p>
                        <Button
                            onClick={() => history.push('/login')}
                            size='medium'
                            emphasis='primary'
                        >
                            <FormattedMessage
                                id='InvitationPage.go-to-login'
                                defaultMessage='Go to Login'
                            />
                        </Button>
                    </div>
                </div>
            </div>
        )
    }

    // Check if user is logged in
    const isLoggedIn = Boolean(client.getLoggedInUser())

    return (
        <div className='InvitationPage'>
            <div className='invitation-container'>
                <div className='invitation-header'>
                    <h1>
                        <FormattedMessage
                            id='InvitationPage.title'
                            defaultMessage='Board Invitation'
                        />
                    </h1>
                </div>

                <div className='invitation-content'>
                    <div className='invitation-details'>
                        <p>
                            <FormattedMessage
                                id='InvitationPage.invited-to'
                                defaultMessage='You have been invited to join:'
                            />
                        </p>
                        <h2 className='board-title'>{invitationInfo.boardTitle}</h2>
                        <p className='invitation-email'>
                            <FormattedMessage
                                id='InvitationPage.for-email'
                                defaultMessage='For: {email}'
                                values={{email: invitationInfo.email}}
                            />
                        </p>
                        <p className='invitation-role'>
                            <FormattedMessage
                                id='InvitationPage.as-role'
                                defaultMessage='Role: {role}'
                                values={{role: invitationInfo.role}}
                            />
                        </p>
                    </div>

                    <div className='invitation-actions'>
                        {isLoggedIn ? (
                            <Button
                                onClick={handleAcceptInvitation}
                                size='medium'
                                emphasis='primary'
                                disabled={accepting}
                            >
                                {accepting ? (
                                    <FormattedMessage
                                        id='InvitationPage.accepting'
                                        defaultMessage='Accepting...'
                                    />
                                ) : (
                                    <FormattedMessage
                                        id='InvitationPage.accept'
                                        defaultMessage='Accept Invitation'
                                    />
                                )}
                            </Button>
                        ) : (
                            <div className='auth-buttons'>
                                <p>
                                    <FormattedMessage
                                        id='InvitationPage.login-required'
                                        defaultMessage='Please log in or create an account to accept this invitation.'
                                    />
                                </p>
                                <div className='button-group'>
                                    <Button
                                        onClick={handleLogin}
                                        size='medium'
                                        emphasis='primary'
                                    >
                                        <FormattedMessage
                                            id='InvitationPage.login'
                                            defaultMessage='Log In'
                                        />
                                    </Button>
                                    <Button
                                        onClick={handleRegister}
                                        size='medium'
                                        emphasis='secondary'
                                    >
                                        <FormattedMessage
                                            id='InvitationPage.register'
                                            defaultMessage='Create Account'
                                        />
                                    </Button>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default InvitationPage