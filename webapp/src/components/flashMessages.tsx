// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
import React, {useState, useEffect} from 'react'
import {createNanoEvents} from 'nanoevents'

import './flashMessages.scss'

export type FlashMessage = {
    content: React.ReactNode
    severity: 'low' | 'normal' | 'high'
    persistent?: boolean
}

const emitter = createNanoEvents()

export function sendFlashMessage(message: FlashMessage): void {
    emitter.emit('message', message)
}

export function clearFlashMessages(): void {
    emitter.emit('clear')
}

type Props = {
    milliseconds: number
}

export const FlashMessages = React.memo((props: Props) => {
    const [message, setMessage] = useState<FlashMessage|null>(null)
    const [fadeOut, setFadeOut] = useState(false)
    const [timeoutId, setTimeoutId] = useState<ReturnType<typeof setTimeout>|null>(null)

    useEffect(() => {
        let isSubscribed = true
        const unsubscribeMessage = emitter.on('message', (newMessage: FlashMessage) => {
            if (isSubscribed) {
                if (timeoutId) {
                    clearTimeout(timeoutId)
                    setTimeoutId(null)
                }
                
                // Only set timeout for non-persistent messages
                if (!newMessage.persistent) {
                    setTimeoutId(setTimeout(handleFadeOut, props.milliseconds - 200))
                }
                setMessage(newMessage)
            }
        })
        
        const unsubscribeClear = emitter.on('clear', () => {
            if (isSubscribed) {
                if (timeoutId) {
                    clearTimeout(timeoutId)
                    setTimeoutId(null)
                }
                handleFadeOut()
            }
        })
        
        return () => {
            isSubscribed = false
            unsubscribeMessage()
            unsubscribeClear()
        }
    }, [])

    const handleFadeOut = (): void => {
        setFadeOut(true)
        setTimeoutId(setTimeout(handleTimeout, 200))
    }

    const handleTimeout = (): void => {
        setMessage(null)
        setFadeOut(false)
    }

    const handleClick = (): void => {
        // Don't allow clicking to dismiss persistent messages
        if (message?.persistent) {
            return
        }
        
        if (timeoutId) {
            clearTimeout(timeoutId)
            setTimeoutId(null)
        }
        handleFadeOut()
    }

    if (!message) {
        return null
    }

    return (
        <div
            className={'FlashMessages ' + message.severity + (fadeOut ? ' flashOut' : ' flashIn') + (message.persistent ? ' persistent' : '')}
            onClick={handleClick}
        >
            {message.content}
        </div>
    )
})
