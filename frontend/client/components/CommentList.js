import React, {useEffect, useState} from "react";
import CommentDisplay from "./CommentDisplay";

const CommentList = ({setCommentStatus, setLikeStatus}) => {
    const [comments, setComments] = useState([]);
    const [likes, setLikes] = useState([]);

    useEffect(() => {
        const wsComment = new WebSocket("ws://localhost:8080/ws/comments");
        const wsLike = new WebSocket("ws://localhost:8080/ws/likes");

        wsComment.onmessage = (event) => {
            const newComments = JSON.parse(event.data);
            setComments(newComments);
        };
        wsLike.onmessage = (event) => {
            const newLikes = JSON.parse(event.data);
            setLikes(newLikes);
        };

        wsComment.onopen = () => {
            console.log("Connected to /comments");
            setCommentStatus("connected");
        };
        wsLike.onopen = () => {
            console.log("Connected to /likes");
            setLikeStatus("connected");
        }

        wsComment.onclose = () => {
            console.log("Disconnected from /comments");
            setCommentStatus("connecting");
        };
        wsLike.onclose = () => {
            console.log("Disconnected from /likes");
            setLikeStatus("connecting");
        };

        wsComment.onerror = (error) => {
            console.error("WebSocket error:", error);
            setCommentStatus("error");
        };
        wsLike.onerror = (error) => {
            console.error("WebSocket error:", error);
            setLikeStatus("error");
        };

        return () => {
            wsComment.close();
            wsLike.close();
        };
    }, []);

    return (
        <div>

            {comments.length > 0 ? comments.map((comment, index) => (
                <CommentDisplay
                    key={index}
                    comment={{
                        avatar: "/images/avatar.jpg", // assuming a static avatar for each comment
                        time: new Date(comment.post_time).toLocaleDateString("en-US", {
                            month: "short",
                            day: "numeric",
                            year: "numeric",
                            hour: "2-digit",
                            minute: "2-digit",
                            hour12: false,
                        }),
                        text: comment.content,
                        emotion: comment.sentiment,
                    }}
                />
            )) : <p>Aucun commentaires disponible</p>}
        </div>
    );
};

export default CommentList;
