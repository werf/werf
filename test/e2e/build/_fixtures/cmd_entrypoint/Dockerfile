# Base Stage without SHELL, CMD and ENTRYPOINT
FROM registry.werf.io/werf/scratch as base

# Base Stage with CMD and without ENTRYPOINT
FROM base as base_with_cmd
CMD echo "CMD (shell, base image)"

# Stage 1: Shell form ENTRYPOINT
FROM base as shell_entrypoint
ENTRYPOINT echo "ENTRYPOINT (shell)"

# Stage 2: Exec form ENTRYPOINT
FROM base as exec_entrypoint
ENTRYPOINT ["echo \"ENTRYPOINT (exec)\""]

# Stage 3: Shell form CMD
FROM base as shell_cmd
CMD echo "CMD (shell)"

# Stage 4: Exec form CMD
FROM base as exec_cmd
CMD ["echo \"CMD (exec)\""]

# Stage 5: No CMD, No ENTRYPOINT
FROM base as no_cmd_no_entrypoint

# Stage 6: Shell form ENTRYPOINT, No CMD (should reset CMD)
FROM base_with_cmd as entrypoint_reset_cmd
ENTRYPOINT echo "ENTRYPOINT (shell)"

# Stage 7: Shell form ENTRYPOINT, No CMD (should reset CMD)
FROM base_with_cmd as entrypoint_cmd
CMD echo "CMD (shell)"
ENTRYPOINT echo "ENTRYPOINT (shell)"

# Stage 8: Base image CMD
FROM base_with_cmd as base_image_cmd

# Stage 9: Custom shell
FROM base_with_cmd as custom_shell_exec_cmd_and_entrypoint
SHELL ["/bin/bash", "-c"]
ENTRYPOINT echo "ENTRYPOINT (shell)"
CMD echo "CMD (shell)"
