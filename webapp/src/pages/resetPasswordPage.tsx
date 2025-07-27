// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useEffect} from 'react'
import {useHistory, useLocation} from 'react-router-dom'
import {FormattedMessage, useIntl} from 'react-intl'

import Button from '../widgets/buttons/button'
import LoadingButton from '../widgets/buttons/loadingButton'
import client from '../octoClient'
import {sendFlashMessage} from '../components/flashMessages'
import './resetPasswordPage.scss'

const ResetPasswordPage = () => {
    const intl = useIntl()
    const history = useHistory()
    const location = useLocation()
    const [password, setPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [token, setToken] = useState('')
    const [isValid, setIsValid] = useState(true)
    const [errors, setErrors] = useState<{password?: string, confirmPassword?: string}>({})

    useEffect(() => {
        const params = new URLSearchParams(location.search)
        const tokenParam = params.get('token')
        if (!tokenParam) {
            setIsValid(false)
        } else {
            setToken(tokenParam)
        }
    }, [location])

    const validateForm = (): boolean => {
        const newErrors: {password?: string, confirmPassword?: string} = {}

        if (!password) {
            newErrors.password = intl.formatMessage({
                id: 'ResetPasswordPage.password-required',
                defaultMessage: 'Password is required',
            })
        } else if (password.length < 8) {
            newErrors.password = intl.formatMessage({
                id: 'ResetPasswordPage.password-min-length',
                defaultMessage: 'Password must be at least 8 characters',
            })
        }

        if (!confirmPassword) {
            newErrors.confirmPassword = intl.formatMessage({
                id: 'ResetPasswordPage.confirm-password-required',
                defaultMessage: 'Please confirm your password',
            })
        } else if (password !== confirmPassword) {
            newErrors.confirmPassword = intl.formatMessage({
                id: 'ResetPasswordPage.passwords-dont-match',
                defaultMessage: 'Passwords do not match',
            })
        }

        setErrors(newErrors)
        return Object.keys(newErrors).length === 0
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (!validateForm()) {
            return
        }

        try {
            await client.resetPassword(token, password)

            sendFlashMessage({
                content: intl.formatMessage({
                    id: 'ResetPasswordPage.success',
                    defaultMessage: 'Password reset successfully! Please log in with your new password.',
                }),
                severity: 'normal',
            })

            history.push('/login')
        } catch (error: any) {
            if (error.message?.includes('expired')) {
                sendFlashMessage({
                    content: intl.formatMessage({
                        id: 'ResetPasswordPage.token-expired',
                        defaultMessage: 'This reset link has expired. Please request a new one.',
                    }),
                    severity: 'high',
                })
            } else if (error.message?.includes('used')) {
                sendFlashMessage({
                    content: intl.formatMessage({
                        id: 'ResetPasswordPage.token-used',
                        defaultMessage: 'This reset link has already been used.',
                    }),
                    severity: 'high',
                })
            } else {
                sendFlashMessage({
                    content: intl.formatMessage({
                        id: 'ResetPasswordPage.error',
                        defaultMessage: 'An error occurred. Please try again.',
                    }),
                    severity: 'high',
                })
            }
        }
    }

    if (!isValid) {
        return (
            <div className='ResetPasswordPage'>
                <div className='reset-password-container'>
                    <div className='reset-password-header'>
                        <h1>
                            <FormattedMessage
                                id='ResetPasswordPage.invalid-link-title'
                                defaultMessage='Invalid Reset Link'
                            />
                        </h1>
                    </div>
                    <div className='reset-password-content'>
                        <p className='error-message'>
                            <FormattedMessage
                                id='ResetPasswordPage.invalid-link-message'
                                defaultMessage='This password reset link is invalid or has expired.'
                            />
                        </p>
                        <Button
                            size='medium'
                            filled={true}
                            onClick={() => history.push('/forgot-password')}
                        >
                            <FormattedMessage
                                id='ResetPasswordPage.request-new-link'
                                defaultMessage='Request New Link'
                            />
                        </Button>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className='ResetPasswordPage'>
            <form
                className='reset-password-container'
                onSubmit={handleSubmit}
            >
                <div className='reset-password-header'>
                    <h1>
                        <FormattedMessage
                            id='ResetPasswordPage.title'
                            defaultMessage='Reset your password'
                        />
                    </h1>
                    <p>
                        <FormattedMessage
                            id='ResetPasswordPage.subtitle'
                            defaultMessage='Enter your new password below'
                        />
                    </p>
                </div>

                <div className='reset-password-content'>
                    <div className='form-group'>
                        <label htmlFor='password'>
                            <FormattedMessage
                                id='ResetPasswordPage.new-password-label'
                                defaultMessage='New Password'
                            />
                        </label>
                        <input
                            id='password'
                            type='password'
                            className={`form-input ${errors.password ? 'error' : ''}`}
                            value={password}
                            onChange={(e) => {
                                setPassword(e.target.value)
                                if (errors.password) {
                                    setErrors({...errors, password: undefined})
                                }
                            }}
                            placeholder={intl.formatMessage({
                                id: 'ResetPasswordPage.new-password-placeholder',
                                defaultMessage: 'Enter new password',
                            })}
                            autoFocus={true}
                            required={true}
                        />
                        {errors.password && (
                            <div className='error-text'>{errors.password}</div>
                        )}
                    </div>

                    <div className='form-group'>
                        <label htmlFor='confirmPassword'>
                            <FormattedMessage
                                id='ResetPasswordPage.confirm-password-label'
                                defaultMessage='Confirm Password'
                            />
                        </label>
                        <input
                            id='confirmPassword'
                            type='password'
                            className={`form-input ${errors.confirmPassword ? 'error' : ''}`}
                            value={confirmPassword}
                            onChange={(e) => {
                                setConfirmPassword(e.target.value)
                                if (errors.confirmPassword) {
                                    setErrors({...errors, confirmPassword: undefined})
                                }
                            }}
                            placeholder={intl.formatMessage({
                                id: 'ResetPasswordPage.confirm-password-placeholder',
                                defaultMessage: 'Confirm new password',
                            })}
                            required={true}
                        />
                        {errors.confirmPassword && (
                            <div className='error-text'>{errors.confirmPassword}</div>
                        )}
                    </div>

                    <LoadingButton
                        submit={true}
                        size='medium'
                        filled={true}
                        loadingText={
                            <FormattedMessage
                                id='ResetPasswordPage.resetting'
                                defaultMessage='Resetting...'
                            />
                        }
                    >
                        <FormattedMessage
                            id='ResetPasswordPage.reset-password-button'
                            defaultMessage='Reset Password'
                        />
                    </LoadingButton>
                </div>
            </form>
        </div>
    )
}

export default ResetPasswordPage
