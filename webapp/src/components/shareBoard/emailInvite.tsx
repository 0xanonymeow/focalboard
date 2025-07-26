// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState} from 'react'
import {FormattedMessage, useIntl} from 'react-intl'

import Button from '../../widgets/buttons/button'
import CompassIcon from '../../widgets/icons/compassIcon'
import {sendFlashMessage} from '../flashMessages'
import client from '../../octoClient'
import {getCurrentBoard} from '../../store/boards'
import {useAppSelector} from '../../store/hooks'

import './emailInvite.scss'

const EmailInvite = () => {
    const intl = useIntl()
    const board = useAppSelector(getCurrentBoard)
    const [email, setEmail] = useState('')
    const [role, setRole] = useState('viewer')
    const [isLoading, setIsLoading] = useState(false)

    const validateEmail = (email: string) => {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
        return re.test(email)
    }

    const handleSendInvite = async () => {
        if (!email.trim()) {
            sendFlashMessage({
                content: intl.formatMessage({
                    id: 'ShareBoard.emailInvite.emailRequired',
                    defaultMessage: 'Email address is required',
                }),
                severity: 'high',
            })
            return
        }

        if (!validateEmail(email)) {
            sendFlashMessage({
                content: intl.formatMessage({
                    id: 'ShareBoard.emailInvite.invalidEmail',
                    defaultMessage: 'Please enter a valid email address',
                }),
                severity: 'high',
            })
            return
        }

        setIsLoading(true)

        try {
            const success = await client.sendBoardInvitation(board.id, email, role)
            
            if (success) {
                sendFlashMessage({
                    content: intl.formatMessage({
                        id: 'ShareBoard.emailInvite.success',
                        defaultMessage: 'Invitation sent successfully',
                    }),
                    severity: 'low',
                })
                setEmail('')
            } else {
                sendFlashMessage({
                    content: intl.formatMessage({
                        id: 'ShareBoard.emailInvite.error',
                        defaultMessage: 'Failed to send invitation. Please try again.',
                    }),
                    severity: 'high',
                })
            }
        } catch (error) {
            sendFlashMessage({
                content: intl.formatMessage({
                    id: 'ShareBoard.emailInvite.error',
                    defaultMessage: 'Failed to send invitation. Please try again.',
                }),
                severity: 'high',
            })
        } finally {
            setIsLoading(false)
        }
    }

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') {
            handleSendInvite()
        }
    }

    return (
        <div className='EmailInvite'>
            <div className='EmailInvite__header'>
                <CompassIcon
                    icon='email-outline'
                    className='EmailInvite__icon'
                />
                <div className='EmailInvite__title'>
                    <FormattedMessage
                        id='ShareBoard.emailInvite.title'
                        defaultMessage='Invite by email'
                    />
                </div>
            </div>
            
            <div className='EmailInvite__form'>
                <div className='EmailInvite__input-group'>
                    <input
                        type='email'
                        className='EmailInvite__email-input'
                        placeholder={intl.formatMessage({
                            id: 'ShareBoard.emailInvite.emailPlaceholder',
                            defaultMessage: 'Enter email address',
                        })}
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        onKeyPress={handleKeyPress}
                        disabled={isLoading}
                    />
                    
                    <select
                        className='EmailInvite__role-select'
                        value={role}
                        onChange={(e) => setRole(e.target.value)}
                        disabled={isLoading}
                    >
                        <option value='viewer'>
                            {intl.formatMessage({
                                id: 'ShareBoard.role.viewer',
                                defaultMessage: 'Viewer',
                            })}
                        </option>
                        <option value='commenter'>
                            {intl.formatMessage({
                                id: 'ShareBoard.role.commenter',
                                defaultMessage: 'Commenter',
                            })}
                        </option>
                        <option value='editor'>
                            {intl.formatMessage({
                                id: 'ShareBoard.role.editor',
                                defaultMessage: 'Editor',
                            })}
                        </option>
                        <option value='admin'>
                            {intl.formatMessage({
                                id: 'ShareBoard.role.admin',
                                defaultMessage: 'Admin',
                            })}
                        </option>
                    </select>
                    
                    <Button
                        emphasis='primary'
                        size='medium'
                        onClick={handleSendInvite}
                        disabled={isLoading}
                        className='EmailInvite__send-button'
                    >
                        {isLoading ? (
                            <FormattedMessage
                                id='ShareBoard.emailInvite.sending'
                                defaultMessage='Sending...'
                            />
                        ) : (
                            <FormattedMessage
                                id='ShareBoard.emailInvite.send'
                                defaultMessage='Send Invitation'
                            />
                        )}
                    </Button>
                </div>
            </div>
            
            <div className='EmailInvite__description'>
                <FormattedMessage
                    id='ShareBoard.emailInvite.description'
                    defaultMessage='Send an email invitation to collaborate on this board'
                />
            </div>
        </div>
    )
}

export default EmailInvite