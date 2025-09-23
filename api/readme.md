    - id: Roles
      type: many-to-many
      ref: Role
      joinTable: app_user_role
      localKeys:
        - field: ID
          join: app_user
      refKeys:
        - field: ID
          join: app_role
