// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState} from 'react'
import {Link} from 'react-router-dom'
import {FormattedMessage, useIntl} from 'react-intl'

import Button from '../widgets/buttons/button'
import LoadingButton from '../widgets/buttons/loadingButton'
import client from '../octoClient'
import {sendFlashMessage} from '../components/flashMessages'
import './forgotPasswordPage.scss'

const ForgotPasswordPage = () => {
    const intl = useIntl()
    const [email, setEmail] = useState('')
    const [submitted, setSubmitted] = useState(false)
    const [cooldownSeconds, setCooldownSeconds] = useState(0)

    React.useEffect(() => {
        if (cooldownSeconds > 0) {
            const timer = setTimeout(() => {
                setCooldownSeconds(cooldownSeconds - 1)
            }, 1000)
            return () => clearTimeout(timer)
        }
    }, [cooldownSeconds])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!email) {
            sendFlashMessage({
                content: intl.formatMessage({
                    id: 'ForgotPasswordPage.email-required',
                    defaultMessage: 'Please enter your email address',
                }),
                severity: 'low',
            })
            return
        }

        try {
            await client.requestPasswordReset(email)
            setSubmitted(true)
        } catch (error: any) {
            if (error.message?.includes('wait')) {
                // Extract seconds from error message
                const match = error.message.match(/wait (\d+) seconds/)
                if (match) {
                    setCooldownSeconds(parseInt(match[1], 10))
                }
                sendFlashMessage({
                    content: error.message,
                    severity: 'low',
                })
            } else {
                sendFlashMessage({
                    content: intl.formatMessage({
                        id: 'ForgotPasswordPage.error',
                        defaultMessage: 'An error occurred. Please try again.',
                    }),
                    severity: 'high',
                })
            }
        }
    }

    if (submitted) {
        return (
            <div className='ForgotPasswordPage'>
                <div className='forgot-password-container'>
                    <div className='forgot-password-header'>
                        <h1>
                            <FormattedMessage
                                id='ForgotPasswordPage.check-email-title'
                                defaultMessage='Check your email'
                            />
                        </h1>
                    </div>
                    <div className='forgot-password-content'>
                        <p className='success-message'>
                            <FormattedMessage
                                id='ForgotPasswordPage.check-email-message'
                                defaultMessage='If an account exists for {email}, you will receive password reset instructions.'
                                values={{email}}
                            />
                        </p>
                        <Link to='/login'>
                            <Button
                                size='medium'
                                emphasis='secondary'
                            >
                                <FormattedMessage
                                    id='ForgotPasswordPage.back-to-login'
                                    defaultMessage='Back to Login'
                                />
                            </Button>
                        </Link>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className='ForgotPasswordPage'>
            <form
                className='forgot-password-container'
                onSubmit={handleSubmit}
            >
                <div className='forgot-password-header'>
                    <h1>
                        <FormattedMessage
                            id='ForgotPasswordPage.title'
                            defaultMessage='Forgot your password?'
                        />
                    </h1>
                    <p>
                        <FormattedMessage
                            id='ForgotPasswordPage.subtitle'
                            defaultMessage="Enter your email and we'll send you a reset link"
                        />
                    </p>
                </div>

                <div className='forgot-password-content'>
                    <div className='form-group'>
                        <label htmlFor='email'>
                            <FormattedMessage
                                id='ForgotPasswordPage.email-label'
                                defaultMessage='Email'
                            />
                        </label>
                        <input
                            id='email'
                            type='email'
                            className='form-input'
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder={intl.formatMessage({
                                id: 'ForgotPasswordPage.email-placeholder',
                                defaultMessage: 'Enter your email',
                            })}
                            autoFocus={true}
                            required={true}
                        />
                    </div>

                    {cooldownSeconds > 0 && (
                        <div className='cooldown-message'>
                            <FormattedMessage
                                id='ForgotPasswordPage.cooldown'
                                defaultMessage='Please wait {seconds} seconds before requesting another reset'
                                values={{seconds: cooldownSeconds}}
                            />
                        </div>
                    )}

                    <LoadingButton
                        submit={true}
                        size='medium'
                        filled={true}
                        disabled={cooldownSeconds > 0}
                        loadingText={
                            <FormattedMessage
                                id='ForgotPasswordPage.sending'
                                defaultMessage='Sending...'
                            />
                        }
                    >
                        <FormattedMessage
                            id='ForgotPasswordPage.send-reset-link'
                            defaultMessage='Send reset link'
                        />
                    </LoadingButton>

                    <div className='form-footer'>
                        <Link to='/login'>
                            <FormattedMessage
                                id='ForgotPasswordPage.back-to-login'
                                defaultMessage='Back to Login'
                            />
                        </Link>
                    </div>
                </div>
            </form>
        </div>
    )
}

export default ForgotPasswordPage
