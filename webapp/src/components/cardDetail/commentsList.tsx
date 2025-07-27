// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
import React, {useState} from 'react'
import {FormattedMessage, useIntl} from 'react-intl'

import {CommentBlock, createCommentBlock} from '../../blocks/commentBlock'
import mutator from '../../mutator'
import {useAppSelector} from '../../store/hooks'
import {Utils} from '../../utils'
import Button from '../../widgets/buttons/button'
import LoadingButton from '../../widgets/buttons/loadingButton'

import {MarkdownEditor} from '../markdownEditor'

import {IUser} from '../../user'
import {getMe} from '../../store/users'
import {useHasCurrentBoardPermissions} from '../../hooks/permissions'
import {Permission} from '../../constants'

import AddCommentTourStep from '../onboardingTour/addComments/addComments'

import Comment from './comment'

import './commentsList.scss'

type Props = {
    comments: readonly CommentBlock[]
    boardId: string
    cardId: string
    readonly: boolean
}

const CommentsList = (props: Props) => {
    const [newComment, setNewComment] = useState('')
    const [isSending, setIsSending] = useState(false)
    const me = useAppSelector<IUser|null>(getMe)
    const canDeleteOthersComments = useHasCurrentBoardPermissions([Permission.DeleteOthersComments])

    const onSendClicked = async () => {
        const commentText = newComment
        if (commentText && !isSending) {
            const {cardId, boardId} = props
            Utils.log(`Send comment: ${commentText}`)
            Utils.assertValue(cardId)

            setIsSending(true)
            try {
                const comment = createCommentBlock()
                comment.parentId = cardId
                comment.boardId = boardId
                comment.title = commentText
                await mutator.insertBlock(boardId, comment, 'add comment')
                setNewComment('')
            } finally {
                setIsSending(false)
            }
        }
    }

    const {comments} = props
    const intl = useIntl()

    const newCommentComponent = (
        <div className='CommentsList__new'>
            <img
                className='comment-avatar'
                src={Utils.getProfilePicture(me?.id)}
            />
            <MarkdownEditor
                className='newcomment'
                text={newComment}
                placeholderText={intl.formatMessage({id: 'CardDetail.new-comment-placeholder', defaultMessage: 'Add a comment...'})}
                onChange={(value: string) => {
                    if (newComment !== value) {
                        setNewComment(value)
                    }
                }}
            />

            {newComment &&
            <LoadingButton
                filled={true}
                onClick={onSendClicked}
                loading={isSending}
                loadingText={
                    <FormattedMessage
                        id='CommentsList.sending'
                        defaultMessage='Sending...'
                    />
                }
            >
                <FormattedMessage
                    id='CommentsList.send'
                    defaultMessage='Send'
                />
            </LoadingButton>
            }

            <AddCommentTourStep/>
        </div>
    )

    return (
        <div className='CommentsList'>
            {/* New comment */}
            {!props.readonly && newCommentComponent}

            {comments.slice(0).reverse().map((comment) => {
                // Only modify _own_ comments, EXCEPT for Admins, which can delete _any_ comment
                // NOTE: editing comments will exist in the future (in addition to deleting)
                const canDeleteComment: boolean = canDeleteOthersComments || me?.id === comment.modifiedBy
                return (
                    <Comment
                        key={comment.id}
                        comment={comment}
                        userImageUrl={Utils.getProfilePicture(comment.modifiedBy)}
                        userId={comment.modifiedBy}
                        readonly={props.readonly || !canDeleteComment}
                    />
                )
            })}

            {/* horizontal divider below comments */}
            {!(comments.length === 0 && props.readonly) && <hr className='CommentsList__divider'/>}
        </div>
    )
}

export default React.memo(CommentsList)
