// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
import React from 'react'

import Button from './button'
import LoadingIcon from '../icons/loadingIcon'

type Props = {
    onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void | Promise<void>
    onMouseOver?: (e: React.MouseEvent<HTMLButtonElement>) => void
    onMouseLeave?: (e: React.MouseEvent<HTMLButtonElement>) => void
    onBlur?: (e: React.FocusEvent<HTMLButtonElement>) => void
    children?: React.ReactNode
    loadingText?: React.ReactNode
    title?: string
    icon?: React.ReactNode
    filled?: boolean
    active?: boolean
    submit?: boolean
    emphasis?: string
    size?: string
    danger?: boolean
    className?: string
    rightIcon?: boolean
    disabled?: boolean
    loading?: boolean
}

function LoadingButton(props: Props): JSX.Element {
    const {loading, loadingText, children, icon, disabled, onClick, ...buttonProps} = props
    const [isLoading, setIsLoading] = React.useState(false)

    const handleClick = async (e: React.MouseEvent<HTMLButtonElement>) => {
        if (!onClick || isLoading || loading) {
            return
        }

        const result = onClick(e)
        if (result && typeof result.then === 'function') {
            setIsLoading(true)
            try {
                await result
            } finally {
                setIsLoading(false)
            }
        }
    }

    const isCurrentlyLoading = loading || isLoading

    return (
        <Button
            {...buttonProps}
            onClick={handleClick}
            disabled={disabled || isCurrentlyLoading}
            icon={isCurrentlyLoading ? <LoadingIcon/> : icon}
        >
            {isCurrentlyLoading ? (loadingText || children) : children}
        </Button>
    )
}

export default React.memo(LoadingButton)