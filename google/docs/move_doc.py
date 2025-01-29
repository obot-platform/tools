def move_doc(drive_service, document_id, folder_path):
    """
    Moves a document to a specified folder path in Google Drive. 
    Creates the folder(s) if they don't exist.
    If folder_path is "/", the document is moved back to the root.

    :param drive_service: Drive API service instance.
    :param document_id: ID of the document to move.
    :param folder_path: Folder path in Google Drive (e.g., "folder1/folder2" or "/").
    :return: None
    """
    if not folder_path or folder_path.strip() == "":
        print("No folder path provided. Skipping move operation.")
        return

    if folder_path.strip() == "/":
        # Move the document back to the root folder
        drive_service.files().update(
            fileId=document_id,
            addParents="root",  # Add to the root folder
            removeParents="root",  # Ensure no redundant updates
            fields="id, parents"
        ).execute()
        print("Document moved back to the root folder.")
        return

    def get_or_create_folder(service, folder_path, parent_id="root"):
        """
        Recursively navigate and create folders in Google Drive.
        :param service: Drive API service instance.
        :param folder_path: Path to the folder (e.g., "folder1/folder2").
        :param parent_id: Parent folder ID to start from (default is root).
        :return: ID of the final folder in the path.
        """
        folder_names = folder_path.strip("/").split("/")
        for folder_name in folder_names:
            # Search for the folder with the current name under the parent ID
            query = f"'{parent_id}' in parents and name = '{folder_name}' and mimeType = 'application/vnd.google-apps.folder' and trashed = false"
            results = service.files().list(q=query, fields="files(id, name)").execute()
            files = results.get('files', [])

            if files:
                # Folder exists, use its ID
                parent_id = files[0]['id']
            else:
                # Folder doesn't exist, create it
                folder_metadata = {
                    "name": folder_name,
                    "mimeType": "application/vnd.google-apps.folder",
                    "parents": [parent_id]
                }
                folder = service.files().create(body=folder_metadata, fields="id").execute()
                parent_id = folder.get("id")

        return parent_id

    # Get or create the target folder in Google Drive
    folder_id = get_or_create_folder(drive_service, folder_path)

    # Move the document to the target folder
    drive_service.files().update(
        fileId=document_id,
        addParents=folder_id,
        removeParents="root",  # Remove from the default root folder
        fields="id, parents"
    ).execute()

    print(f"Document with ID {document_id} moved to folder: {folder_path}")
