ALTER TABLE pullreq_activities
    DROP COLUMN pullreq_activity_outdated,
    DROP COLUMN pullreq_activity_code_comment_merge_base_sha,
    DROP COLUMN pullreq_activity_code_comment_source_sha,
    DROP COLUMN pullreq_activity_code_comment_path,
    DROP COLUMN pullreq_activity_code_comment_line_new,
    DROP COLUMN pullreq_activity_code_comment_span_new,
    DROP COLUMN pullreq_activity_code_comment_line_old,
    DROP COLUMN pullreq_activity_code_comment_span_old;
